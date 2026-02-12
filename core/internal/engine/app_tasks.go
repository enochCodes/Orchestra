package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/enochcodes/orchestra/core/internal/buildpack"
	"github.com/enochcodes/orchestra/core/internal/model"
	sshpkg "github.com/enochcodes/orchestra/core/pkg/ssh"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

const (
	TypeDeployApplication = "app:deploy"
)

type DeployAppPayload struct {
	AppID uint `json:"app_id"`
}

type AppTaskHandler struct {
	DB            *gorm.DB
	EncryptionKey string
}

func NewDeployAppTask(appID uint) (*asynq.Task, error) {
	payload, err := json.Marshal(DeployAppPayload{AppID: appID})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeDeployApplication, payload, asynq.Queue("deployment"), asynq.MaxRetry(2)), nil
}

func (h *AppTaskHandler) HandleDeployAppTask(ctx context.Context, t *asynq.Task) error {
	var p DeployAppPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	log.Printf("Starting deployment for App ID: %d", p.AppID)

	var app model.Application
	if err := h.DB.Preload("Cluster").Preload("Cluster.ManagerServer").First(&app, p.AppID).Error; err != nil {
		return fmt.Errorf("app lookup failed: %v", err)
	}

	// Count existing deployments for versioning
	var count int64
	h.DB.Model(&model.Deployment{}).Where("application_id = ?", app.ID).Count(&count)
	version := fmt.Sprintf("v%d", count+1)

	// Create deployment record
	deployment := model.Deployment{
		ApplicationID: app.ID,
		Version:       version,
		Status:        model.DeploymentStatusBuilding,
	}
	if err := h.DB.Create(&deployment).Error; err != nil {
		return fmt.Errorf("create deployment: %v", err)
	}

	h.DB.Model(&app).Update("status", "building")

	// Get SSH client to the manager server
	managerServer := app.Cluster.ManagerServer
	sshKey, err := decrypt(managerServer.SSHKeyEncrypted, h.EncryptionKey)
	if err != nil {
		h.failDeployment(&deployment, &app, "Failed to decrypt manager SSH key")
		return fmt.Errorf("decrypt SSH key: %w", err)
	}

	client, err := sshpkg.NewClient(managerServer.IP, managerServer.SSHPort, managerServer.SSHUser, sshKey, "")
	if err != nil {
		h.failDeployment(&deployment, &app, fmt.Sprintf("SSH to manager failed: %v", err))
		return fmt.Errorf("SSH to manager: %w", err)
	}
	defer client.Close()

	appDir := fmt.Sprintf("/opt/orchestra/apps/%s", sanitizeName(app.Name))
	imageName := fmt.Sprintf("orchestra/%s:%s", sanitizeName(app.Name), version)

	// Step 1: Prepare app directory
	client.ExecuteCommand(fmt.Sprintf("mkdir -p %s", appDir))

	// Step 2: Get source code based on source type
	switch app.SourceType {
	case model.DeploymentSourceGit:
		h.appendLog(&deployment, fmt.Sprintf("Cloning %s (branch: %s)...", app.RepoURL, app.Branch))
		cloneCmd := fmt.Sprintf("cd %s && rm -rf src && git clone --depth 1 --branch %s %s src 2>&1",
			appDir, app.Branch, app.RepoURL)
		result, err := client.ExecuteCommand(cloneCmd)
		if err != nil {
			h.failDeployment(&deployment, &app, fmt.Sprintf("Git clone failed: %s", result.Stderr))
			return fmt.Errorf("git clone: %w", err)
		}
		h.appendLog(&deployment, "Clone complete.")

	case model.DeploymentSourceDocker:
		// Docker image: just pull and deploy directly
		h.appendLog(&deployment, fmt.Sprintf("Pulling Docker image: %s", app.DockerImage))
		imageName = app.DockerImage
		pullCmd := fmt.Sprintf("docker pull %s 2>&1", app.DockerImage)
		result, err := client.ExecuteCommand(pullCmd)
		if err != nil {
			h.failDeployment(&deployment, &app, fmt.Sprintf("Docker pull failed: %s", result.Stderr))
			return fmt.Errorf("docker pull: %w", err)
		}
		h.appendLog(&deployment, "Pull complete.")

	case model.DeploymentSourceManual:
		h.appendLog(&deployment, fmt.Sprintf("Using manual path: %s", app.ManualPath))
		client.ExecuteCommand(fmt.Sprintf("cd %s && ln -sfn %s src", appDir, app.ManualPath))
	}

	// Step 3: Build (if not docker_image source)
	if app.SourceType != model.DeploymentSourceDocker {
		srcDir := fmt.Sprintf("%s/src", appDir)

		// Check if repo has Dockerfile
		hasDockerfile := false
		checkResult, _ := client.ExecuteCommand(fmt.Sprintf("test -f %s/Dockerfile && echo YES || echo NO", srcDir))
		if strings.TrimSpace(checkResult.Stdout) == "YES" {
			hasDockerfile = true
		}

		if !hasDockerfile && app.BuildType != "" && app.BuildType != "docker" {
			// Generate Dockerfile from buildpack
			dockerfile := buildpack.GenerateDockerfile(app.BuildType, app.BuildCmd, app.StartCmd)
			if dockerfile != "" {
				h.appendLog(&deployment, "Generating Dockerfile from buildpack...")
				escapedDF := strings.ReplaceAll(dockerfile, "'", "'\\''")
				writeCmd := fmt.Sprintf("cat > %s/Dockerfile << 'ORCHESTRA_EOF'\n%s\nORCHESTRA_EOF", srcDir, escapedDF)
				client.ExecuteCommand(writeCmd)
			}
		}

		h.appendLog(&deployment, "Building Docker image...")
		h.DB.Model(&deployment).Update("status", model.DeploymentStatusBuilding)
		buildCmd := fmt.Sprintf("cd %s && docker build -t %s . 2>&1", srcDir, imageName)
		result, err := client.ExecuteCommand(buildCmd)
		if err != nil {
			h.failDeployment(&deployment, &app, fmt.Sprintf("Docker build failed: %s", result.Stderr))
			return fmt.Errorf("docker build: %w", err)
		}
		h.appendLog(&deployment, "Build complete.")
	}

	// Step 4: Deploy based on cluster type
	h.DB.Model(&deployment).Update("status", model.DeploymentStatusDeploying)
	h.DB.Model(&app).Update("status", "deploying")

	// Build env args
	envArgs := h.buildEnvArgs(app)
	containerName := sanitizeName(app.Name)
	portMapping := ""
	if app.Port > 0 {
		portMapping = fmt.Sprintf("-p %d:%d", app.Port, app.Port)
	}

	switch app.Cluster.Type {
	case model.ClusterTypeK8s:
		err = h.deployK8s(client, &deployment, &app, imageName, containerName, envArgs)
	case model.ClusterTypeDockerSwarm:
		err = h.deploySwarm(client, &deployment, &app, imageName, containerName, envArgs, portMapping)
	case model.ClusterTypeManual:
		err = h.deployDocker(client, &deployment, &app, imageName, containerName, envArgs, portMapping)
	default:
		err = h.deployDocker(client, &deployment, &app, imageName, containerName, envArgs, portMapping)
	}

	if err != nil {
		return err
	}

	// Step 5: Mark as live
	h.DB.Model(&deployment).Updates(map[string]interface{}{
		"status":    model.DeploymentStatusLive,
		"image_tag": imageName,
	})
	h.DB.Model(&app).Update("status", "running")
	h.appendLog(&deployment, fmt.Sprintf("Deployment %s is live!", version))

	log.Printf("Deployment %s complete for app %s", version, app.Name)
	return nil
}

func (h *AppTaskHandler) deployK8s(client *sshpkg.Client, dep *model.Deployment, app *model.Application, image, name, envArgs string) error {
	h.appendLog(dep, "Deploying to Kubernetes...")

	// Generate K8s manifest
	envYaml := ""
	if app.EnvVars.Production != nil {
		envYaml = "        env:\n"
		for k, v := range app.EnvVars.Production {
			envYaml += fmt.Sprintf("        - name: %s\n          value: \"%s\"\n", k, v)
		}
	}

	portYaml := ""
	if app.Port > 0 {
		portYaml = fmt.Sprintf("        ports:\n        - containerPort: %d", app.Port)
	}

	manifest := fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: %s
  namespace: %s
spec:
  replicas: %d
  selector:
    matchLabels:
      app: %s
  template:
    metadata:
      labels:
        app: %s
    spec:
      containers:
      - name: %s
        image: %s
%s%s
---
apiVersion: v1
kind: Service
metadata:
  name: %s
  namespace: %s
spec:
  selector:
    app: %s
  ports:
  - port: %d
    targetPort: %d
  type: ClusterIP`,
		name, app.Namespace, app.Replicas, name, name, name, image,
		envYaml, portYaml,
		name, app.Namespace, name,
		app.Port, app.Port,
	)

	// Write and apply manifest
	writeCmd := fmt.Sprintf("cat > /tmp/%s.yaml << 'ORCHESTRA_EOF'\n%s\nORCHESTRA_EOF", name, manifest)
	client.ExecuteCommand(writeCmd)

	result, err := client.ExecuteCommand(fmt.Sprintf("kubectl apply -f /tmp/%s.yaml 2>&1", name))
	if err != nil {
		h.failDeployment(dep, app, fmt.Sprintf("kubectl apply failed: %s", result.Stderr))
		return fmt.Errorf("kubectl apply: %w", err)
	}
	h.appendLog(dep, "Kubernetes deployment applied.")
	return nil
}

func (h *AppTaskHandler) deploySwarm(client *sshpkg.Client, dep *model.Deployment, app *model.Application, image, name, envArgs, portMapping string) error {
	h.appendLog(dep, "Deploying to Docker Swarm...")

	// Remove existing service
	client.ExecuteCommand(fmt.Sprintf("docker service rm %s 2>/dev/null", name))

	cmd := fmt.Sprintf("docker service create --name %s --replicas %d %s %s %s 2>&1",
		name, app.Replicas, envArgs, portMapping, image)
	result, err := client.ExecuteCommand(cmd)
	if err != nil {
		h.failDeployment(dep, app, fmt.Sprintf("Swarm deploy failed: %s", result.Stderr))
		return fmt.Errorf("swarm deploy: %w", err)
	}
	h.appendLog(dep, "Swarm service created.")
	return nil
}

func (h *AppTaskHandler) deployDocker(client *sshpkg.Client, dep *model.Deployment, app *model.Application, image, name, envArgs, portMapping string) error {
	h.appendLog(dep, "Deploying with Docker...")

	// Stop existing container
	client.ExecuteCommand(fmt.Sprintf("docker stop %s 2>/dev/null; docker rm %s 2>/dev/null", name, name))

	cmd := fmt.Sprintf("docker run -d --name %s --restart unless-stopped %s %s %s 2>&1",
		name, envArgs, portMapping, image)
	result, err := client.ExecuteCommand(cmd)
	if err != nil {
		h.failDeployment(dep, app, fmt.Sprintf("Docker run failed: %s", result.Stderr))
		return fmt.Errorf("docker run: %w", err)
	}
	h.appendLog(dep, "Docker container started.")
	return nil
}

func (h *AppTaskHandler) buildEnvArgs(app model.Application) string {
	var parts []string
	if app.EnvVars.Production != nil {
		for k, v := range app.EnvVars.Production {
			parts = append(parts, fmt.Sprintf("-e %s='%s'", k, v))
		}
	}
	return strings.Join(parts, " ")
}

func (h *AppTaskHandler) failDeployment(dep *model.Deployment, app *model.Application, msg string) {
	h.appendLog(dep, fmt.Sprintf("ERROR: %s", msg))
	h.DB.Model(dep).Update("status", model.DeploymentStatusFailed)
	h.DB.Model(app).Update("status", "failed")
}

func (h *AppTaskHandler) appendLog(dep *model.Deployment, line string) {
	h.DB.Model(dep).Update("logs", gorm.Expr("COALESCE(logs, '') || ?", line+"\n"))
}

func sanitizeName(name string) string {
	r := strings.NewReplacer(" ", "-", "_", "-", ".", "-")
	return strings.ToLower(r.Replace(name))
}
