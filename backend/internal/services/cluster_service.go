package services

import (
	"fmt"
	"log"

	tasks "github.com/enochcodes/orchestra/backend/internal/engine"
	"github.com/enochcodes/orchestra/backend/internal/models"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

// ClusterService provides business logic for cluster operations.
type ClusterService struct {
	DB            *gorm.DB
	AsynqClient   *asynq.Client
	EncryptionKey string
}

// NewClusterService creates a new ClusterService.
func NewClusterService(db *gorm.DB, client *asynq.Client, encryptionKey string) *ClusterService {
	return &ClusterService{
		DB:            db,
		AsynqClient:   client,
		EncryptionKey: encryptionKey,
	}
}

// DesignClusterInput represents the input for designing a new cluster.
type DesignClusterInput struct {
	Name            string `json:"name" validate:"required"`
	Type            string `json:"type"` // k8s, docker_swarm, manual
	ManagerServerID uint   `json:"manager_server_id" validate:"required"`
	WorkerServerIDs []uint `json:"worker_server_ids"`
	CNIPlugin       string `json:"cni_plugin"`
	Domain          string `json:"domain"`
}

// DesignCluster creates a new cluster and enqueues provisioning tasks based on cluster type.
func (s *ClusterService) DesignCluster(input DesignClusterInput) (*models.Cluster, error) {
	// Validate manager server
	var manager models.Server
	if err := s.DB.First(&manager, input.ManagerServerID).Error; err != nil {
		return nil, fmt.Errorf("manager server not found: %w", err)
	}
	if manager.Status != models.ServerStatusReady {
		return nil, fmt.Errorf("manager server is not in 'ready' state (current: %s)", manager.Status)
	}

	// Defaults
	clusterType := models.ClusterType(input.Type)
	if clusterType == "" {
		clusterType = models.ClusterTypeK8s
	}

	cni := input.CNIPlugin
	if cni == "" && clusterType == models.ClusterTypeK8s {
		cni = "flannel"
	}

	// Create cluster
	cluster := models.Cluster{
		Name:            input.Name,
		Type:            clusterType,
		ManagerServerID: input.ManagerServerID,
		CNIPlugin:       cni,
		Domain:          input.Domain,
		Status:          models.ClusterStatusPending,
	}
	if err := s.DB.Create(&cluster).Error; err != nil {
		return nil, fmt.Errorf("failed to create cluster: %w", err)
	}

	// Enqueue provisioning based on cluster type
	switch clusterType {
	case models.ClusterTypeK8s:
		if err := s.provisionK8s(cluster.ID, input); err != nil {
			return &cluster, err
		}
	case models.ClusterTypeDockerSwarm:
		if err := s.provisionSwarm(cluster.ID, input); err != nil {
			return &cluster, err
		}
	case models.ClusterTypeManual:
		if err := s.provisionManual(cluster.ID, input); err != nil {
			return &cluster, err
		}
	default:
		return nil, fmt.Errorf("unsupported cluster type: %s", clusterType)
	}

	return &cluster, nil
}

func (s *ClusterService) provisionK8s(clusterID uint, input DesignClusterInput) error {
	task, err := tasks.NewDesignateManagerTask(clusterID, input.ManagerServerID)
	if err != nil {
		return fmt.Errorf("create manager task: %w", err)
	}
	if _, err := s.AsynqClient.Enqueue(task); err != nil {
		return fmt.Errorf("enqueue manager task: %w", err)
	}
	log.Printf("Enqueued K3s manager task for cluster %d", clusterID)

	for _, workerID := range input.WorkerServerIDs {
		t, err := tasks.NewJoinWorkerTask(clusterID, workerID)
		if err != nil {
			log.Printf("Failed to create K8s worker task for %d: %v", workerID, err)
			continue
		}
		if _, err := s.AsynqClient.Enqueue(t); err != nil {
			log.Printf("Failed to enqueue K8s worker task for %d: %v", workerID, err)
		}
	}
	return nil
}

func (s *ClusterService) provisionSwarm(clusterID uint, input DesignClusterInput) error {
	task, err := tasks.NewSwarmInitTask(clusterID, input.ManagerServerID)
	if err != nil {
		return fmt.Errorf("create swarm init task: %w", err)
	}
	if _, err := s.AsynqClient.Enqueue(task); err != nil {
		return fmt.Errorf("enqueue swarm init task: %w", err)
	}
	log.Printf("Enqueued Swarm init task for cluster %d", clusterID)

	for _, workerID := range input.WorkerServerIDs {
		t, err := tasks.NewSwarmJoinTask(clusterID, workerID)
		if err != nil {
			log.Printf("Failed to create Swarm join task for %d: %v", workerID, err)
			continue
		}
		if _, err := s.AsynqClient.Enqueue(t); err != nil {
			log.Printf("Failed to enqueue Swarm join task for %d: %v", workerID, err)
		}
	}
	return nil
}

func (s *ClusterService) provisionManual(clusterID uint, input DesignClusterInput) error {
	// Manual clusters: install Docker on all nodes, mark cluster active
	task, err := tasks.NewManualClusterSetupTask(clusterID, input.ManagerServerID, input.WorkerServerIDs)
	if err != nil {
		return fmt.Errorf("create manual setup task: %w", err)
	}
	if _, err := s.AsynqClient.Enqueue(task); err != nil {
		return fmt.Errorf("enqueue manual setup task: %w", err)
	}
	log.Printf("Enqueued manual cluster setup for cluster %d", clusterID)
	return nil
}

// GetCluster retrieves a cluster by ID with its relations.
func (s *ClusterService) GetCluster(id uint) (*models.Cluster, error) {
	var cluster models.Cluster
	if err := s.DB.Preload("ManagerServer").Preload("Workers").First(&cluster, id).Error; err != nil {
		return nil, fmt.Errorf("cluster not found: %w", err)
	}
	return &cluster, nil
}

// ListClusters retrieves all clusters.
func (s *ClusterService) ListClusters() ([]models.Cluster, error) {
	var clusters []models.Cluster
	if err := s.DB.Preload("ManagerServer").Preload("Workers").Find(&clusters).Error; err != nil {
		return nil, fmt.Errorf("failed to list clusters: %w", err)
	}
	return clusters, nil
}
