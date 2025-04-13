package usecase_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/aferryc/yars/internal/config"
	"github.com/aferryc/yars/model"
	repositorymock "github.com/aferryc/yars/repository/mocks"
	"github.com/aferryc/yars/usecase"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestReconManager_GenerateUploadURLs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGCSRepo := repositorymock.NewMockGCSRepository(ctrl)
	mockKafkaRepo := repositorymock.NewMockKafkaRepository(ctrl)

	cfg := &config.Config{
		Kafka: config.KafkaConfig{
			Topic: config.TopicConfig{
				CompilerTopic: "test-compiler-topic",
			},
		},
	}

	manager := usecase.NewReconManager(mockGCSRepo, mockKafkaRepo, cfg)

	t.Run("Successfully generate upload URLs", func(t *testing.T) {
		// Setup expectations
		expectedTransactionURL := "https://storage.googleapis.com/bucket/uploads/uuid/transactions.csv"
		expectedBankStatementURL := "https://storage.googleapis.com/bucket/uploads/uuid/bank_statement.csv"

		// Use a matcher for the taskID since it's a UUID we can't predict
		mockGCSRepo.EXPECT().
			GenerateUploadURL(gomock.Any(), "text/csv", gomock.Any()).
			DoAndReturn(func(path, contentType string, expires time.Time) (string, error) {
				// Ensure path starts with "uploads/" and contains the file name
				assert.Contains(t, path, "uploads/")
				if path == "" {
					return "", errors.New("invalid path")
				}

				// Return different URLs based on the path
				if path == "" {
					return "", errors.New("path cannot be empty")
				}
				// Use regex to determine which file we're generating a URL for
				if strings.Contains(path, "transactions.csv") {
					return expectedTransactionURL, nil
				} else if strings.Contains(path, "bank_statement.csv") {
					return expectedBankStatementURL, nil
				}
				return expectedTransactionURL, nil
			}).
			Times(2)

		// Call the method
		ctx := context.Background()
		response, err := manager.GenerateUploadURLs(ctx)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotEmpty(t, response.TaskID)
		assert.Equal(t, expectedTransactionURL, response.TransactionURL)
		assert.Equal(t, expectedBankStatementURL, response.BankStatementURL)
		assert.True(t, response.ExpiresAt.After(time.Now()))

		// Validate UUID format
		_, err = uuid.Parse(response.TaskID)
		assert.NoError(t, err, "TaskID should be a valid UUID")
	})

	t.Run("Error generating transaction URL", func(t *testing.T) {
		// Setup expectations
		mockGCSRepo.EXPECT().
			GenerateUploadURL(gomock.Any(), "text/csv", gomock.Any()).
			Return("", errors.New("GCS error")).
			Times(1)

		// Call the method
		ctx := context.Background()
		response, err := manager.GenerateUploadURLs(ctx)

		fmt.Println(response)
		// Assert
		assert.Error(t, err)
		assert.Nil(t, response)
	})

	t.Run("Error generating bank statement URL", func(t *testing.T) {
		// Setup expectations - first call succeeds, second fails
		gomock.InOrder(
			mockGCSRepo.EXPECT().
				GenerateUploadURL(gomock.Any(), "text/csv", gomock.Any()).
				Return("transaction-url", nil),
			mockGCSRepo.EXPECT().
				GenerateUploadURL(gomock.Any(), "text/csv", gomock.Any()).
				Return("", errors.New("GCS error")),
		)

		// Call the method
		ctx := context.Background()
		response, err := manager.GenerateUploadURLs(ctx)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, response)
	})
}

func TestReconManager_InitiateCompilation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGCSRepo := repositorymock.NewMockGCSRepository(ctrl)
	mockKafkaRepo := repositorymock.NewMockKafkaRepository(ctrl)

	cfg := &config.Config{
		Kafka: config.KafkaConfig{
			Topic: config.TopicConfig{
				CompilerTopic: "test-compiler-topic",
			},
		},
	}

	manager := usecase.NewReconManager(mockGCSRepo, mockKafkaRepo, cfg)
	ctx := context.Background()

	t.Run("Successfully initiate compilation", func(t *testing.T) {
		// Setup test data
		taskID := uuid.New().String()
		bankName := "Test Bank"
		startDate := time.Now().Add(-24 * time.Hour)
		endDate := time.Now()

		req := model.CompilerRequest{
			TaskID:    taskID,
			BankName:  bankName,
			StartDate: startDate,
			EndDate:   endDate,
		}

		// Setup expectations
		mockKafkaRepo.EXPECT().
			Publish(
				ctx,
				cfg.Kafka.Topic.CompilerTopic,
				taskID,
				gomock.Any(),
			).
			DoAndReturn(func(ctx context.Context, topic, key string, message any) error {
				// Validate the event structure
				event, ok := message.(model.CompilerEvent)
				if !ok {
					return errors.New("message is not a CompilerEvent")
				}

				// Verify event fields
				assert.Equal(t, taskID, event.TaskID)
				assert.Equal(t, bankName, event.BankName)
				assert.Equal(t, startDate, event.StartDate)
				assert.Equal(t, endDate, event.EndDate)

				// Check file paths
				assert.Contains(t, event.Transaction, "uploads/")
				assert.Contains(t, event.Transaction, taskID)
				assert.Contains(t, event.Transaction, "transactions.csv")

				assert.Contains(t, event.BankStatement, "uploads/")
				assert.Contains(t, event.BankStatement, taskID)
				assert.Contains(t, event.BankStatement, "bank_statement.csv")

				return nil
			})

		// Call the method
		err := manager.InitiateCompilation(ctx, req)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("Missing TaskID", func(t *testing.T) {
		req := model.CompilerRequest{
			BankName: "Test Bank",
		}

		err := manager.InitiateCompilation(ctx, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "task ID is required")
	})

	t.Run("Missing BankName", func(t *testing.T) {
		req := model.CompilerRequest{
			TaskID: uuid.New().String(),
		}

		err := manager.InitiateCompilation(ctx, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bank name is required")
	})

	t.Run("Kafka publish error", func(t *testing.T) {
		req := model.CompilerRequest{
			TaskID:   uuid.New().String(),
			BankName: "Test Bank",
		}

		// Setup expectations
		mockKafkaRepo.EXPECT().
			Publish(ctx, cfg.Kafka.Topic.CompilerTopic, req.TaskID, gomock.Any()).
			Return(errors.New("kafka error"))

		err := manager.InitiateCompilation(ctx, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to publish compilation event")
	})
}

// TestReconManager_Helpers tests the helper functions for directory path generation
func TestReconManager_Helpers(t *testing.T) {
	// This test uses the fact that the path formatting functions are exported indirectly
	// through the ReconManager's behavior

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGCSRepo := repositorymock.NewMockGCSRepository(ctrl)
	mockKafkaRepo := repositorymock.NewMockKafkaRepository(ctrl)

	cfg := &config.Config{
		Kafka: config.KafkaConfig{
			Topic: config.TopicConfig{
				CompilerTopic: "test-compiler-topic",
			},
		},
	}

	manager := usecase.NewReconManager(mockGCSRepo, mockKafkaRepo, cfg)

	// Call InitiateCompilation to test the path formation indirectly
	taskID := "test-uuid"
	req := model.CompilerRequest{
		TaskID:   taskID,
		BankName: "Test Bank",
	}

	// Set up expectations to capture the paths
	var capturedTransaction, capturedBankStatement string
	mockKafkaRepo.EXPECT().
		Publish(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, topic, key string, message any) error {
			event := message.(model.CompilerEvent)
			capturedTransaction = event.Transaction
			capturedBankStatement = event.BankStatement
			return nil
		})

	// Call the method
	_ = manager.InitiateCompilation(context.Background(), req)

	// Assert the path formats
	expectedTransactionPath := "uploads/" + taskID + "/transactions.csv"
	expectedBankStatementPath := "uploads/" + taskID + "/bank_statement.csv"

	assert.Equal(t, expectedTransactionPath, capturedTransaction)
	assert.Equal(t, expectedBankStatementPath, capturedBankStatement)
}
