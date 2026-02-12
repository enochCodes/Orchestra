package services

import (
	"fmt"
	"log"

	"github.com/enochcodes/orchestra/backend/internal/models"
	"github.com/enochcodes/orchestra/backend/internal/tasks"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

// Note: Activity logging for cluster events is done in the task handlers.

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
	ManagerServerID uint   `json:"manager_server_id" validate:"required"`
	WorkerServerIDs []uint `json:"worker_server_ids"`
	CNIPlugin       string `json:"cni_plugin"`
}

// DesignCluster creates a new cluster, designates a manager, and enqueues worker join tasks.
func (s *ClusterService) DesignCluster(input DesignClusterInput) (*models.Cluster, error) {
	// Validate manager server exists and is ready
	var manager models.Server
	if err := s.DB.First(&manager, input.ManagerServerID).Error; err != nil {
		return nil, fmt.Errorf("manager server not found: %w", err)
	}
	if manager.Status != models.ServerStatusReady {
		return nil, fmt.Errorf("manager server is not in 'ready' state (current: %s)", manager.Status)
	}

	// Set default CNI
	cni := input.CNIPlugin
	if cni == "" {
		cni = "flannel"
	}

	// Create cluster record
	cluster := models.Cluster{
		Name:            input.Name,
		ManagerServerID: input.ManagerServerID,
		CNIPlugin:       cni,
		Status:          models.ClusterStatusPending,
	}
	if err := s.DB.Create(&cluster).Error; err != nil {
		return nil, fmt.Errorf("failed to create cluster: %w", err)
	}

	// Enqueue manager designation task
	task, err := tasks.NewDesignateManagerTask(cluster.ID, input.ManagerServerID)
	if err != nil {
		return nil, fmt.Errorf("failed to create manager task: %w", err)
	}
	if _, err := s.AsynqClient.Enqueue(task); err != nil {
		return nil, fmt.Errorf("failed to enqueue manager task: %w", err)
	}
	log.Printf("Enqueued manager designation task for cluster %d", cluster.ID)

	// Enqueue worker join tasks (these will wait for the manager to be ready via retry logic)
	for _, workerID := range input.WorkerServerIDs {
		task, err := tasks.NewJoinWorkerTask(cluster.ID, workerID)
		if err != nil {
			log.Printf("Failed to create join task for worker %d: %v", workerID, err)
			continue
		}
		if _, err := s.AsynqClient.Enqueue(task); err != nil {
			log.Printf("Failed to enqueue join task for worker %d: %v", workerID, err)
			continue
		}
		log.Printf("Enqueued worker join task for server %d -> cluster %d", workerID, cluster.ID)
	}

	return &cluster, nil
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
	if err := s.DB.Preload("ManagerServer").Find(&clusters).Error; err != nil {
		return nil, fmt.Errorf("failed to list clusters: %w", err)
	}
	return clusters, nil
}
