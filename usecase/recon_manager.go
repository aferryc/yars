package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/aferryc/yars/internal/config"
	"github.com/aferryc/yars/model"
	"github.com/aferryc/yars/repository"
	"github.com/google/uuid"
)

const fileDirFormat = "uploads/%s/%s"

type ReconManager struct {
	gcsRepo   repository.GCSRepository
	kafkaRepo repository.KafkaRepository
	cfg       *config.Config
}

func NewReconManager(
	gcsRepo repository.GCSRepository,
	kafkaRepo repository.KafkaRepository,
	cfg *config.Config,
) *ReconManager {
	return &ReconManager{
		gcsRepo:   gcsRepo,
		kafkaRepo: kafkaRepo,
		cfg:       cfg,
	}
}

func (rm *ReconManager) GenerateUploadURLs(ctx context.Context) (*model.UploadURLResponse, error) {
	taskID := uuid.New().String()

	transactionPath := transactionDirectory(taskID)
	bankStatementPath := bankDirectory(taskID)

	expires := time.Now().Add(15 * time.Minute)

	transactionURL, err := rm.gcsRepo.GenerateUploadURL(transactionPath, "text/csv", expires)
	if err != nil {
		return nil, fmt.Errorf("failed to generate transaction upload URL: %w", err)
	}

	bankStatementURL, err := rm.gcsRepo.GenerateUploadURL(bankStatementPath, "text/csv", expires)
	if err != nil {
		return nil, fmt.Errorf("failed to generate bank statement upload URL: %w", err)
	}

	return &model.UploadURLResponse{
		TransactionURL:   transactionURL,
		BankStatementURL: bankStatementURL,
		TaskID:           taskID,
		ExpiresAt:        expires,
	}, nil
}

func (rm *ReconManager) InitiateCompilation(ctx context.Context, req model.CompilerRequest) error {
	if req.TaskID == "" {
		return fmt.Errorf("task ID is required")
	}

	if req.BankName == "" {
		return fmt.Errorf("bank name is required")
	}

	transactionPath := transactionDirectory(req.TaskID)
	bankStatementPath := bankDirectory(req.TaskID)

	event := model.CompilerEvent{
		Transaction:   transactionPath,
		BankStatement: bankStatementPath,
		BankName:      req.BankName,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		TaskID:        req.TaskID,
	}

	err := rm.kafkaRepo.Publish(ctx, rm.cfg.Kafka.Topic.CompilerTopic, req.TaskID, event)
	if err != nil {
		return fmt.Errorf("failed to publish compilation event: %w", err)
	}

	return nil
}

func bankDirectory(taskID string) string {
	return fmt.Sprintf(fileDirFormat, taskID, model.BankStatementFile)
}
func transactionDirectory(taskID string) string {
	return fmt.Sprintf(fileDirFormat, taskID, model.TransactionFile)
}
