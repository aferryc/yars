package usecase_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aferryc/yars/internal/config"
	"github.com/aferryc/yars/model"
	repositorymock "github.com/aferryc/yars/repository/mocks"
	"github.com/aferryc/yars/usecase"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

const transactionFile = "uploads/test-task-id/transactions.csv"
const bankStatementFile = "uploads/test-task-id/bank_statement.csv"

// TestFileCompilerProcessEvent tests the ProcessEvent method of FileCompiler
func TestFileCompilerProcessEvent(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "compiler-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Test cases
	tests := []struct {
		name           string
		event          model.CompilerEvent
		fileContent    string
		setupMocks     func(*testing.T, *mockFileSetup, string)
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name: "Process internal transactions successfully",
			event: model.CompilerEvent{
				Transaction:   transactionFile,
				BankStatement: bankStatementFile,
				TaskID:        "test-task-id",
				BankName:      "TestBank",
				StartDate:     time.Now().Add(-24 * time.Hour),
				EndDate:       time.Now(),
			},
			fileContent: "id,amount,type,timestamp\ntx123,100.50,CREDIT,2023-01-15T14:30:45Z\ntx456,200.75,DEBIT,2023-01-16T10:20:30Z",
			setupMocks: func(t *testing.T, m *mockFileSetup, filePath string) {
				// Mock GCS download for both files
				m.gcsRepo.EXPECT().
					DownloadFromBucket(gomock.Any(), transactionFile).
					Return(createTempFileWithContent(t, filePath), nil)

				m.gcsRepo.EXPECT().
					DownloadFromBucket(gomock.Any(), bankStatementFile).
					Return(createTempFileWithContent(t, "id,amount,date\n101,500.25,2023-01-15\n102,750.50,2023-01-16"), nil)

				// Expect save calls for each transaction
				m.txRepo.EXPECT().Save(gomock.Any()).Return(nil).Times(2)

				// Expect save calls for each bank statement
				m.bankStmtRepo.EXPECT().Save(gomock.Any()).Return(nil).Times(2)

				// Expect Kafka publish call to trigger reconciliation
				m.kafkaRepo.EXPECT().
					Publish(
						gomock.Any(),
						gomock.Any(),
						"test-task-id",
						gomock.Any(),
					).DoAndReturn(func(ctx any, topic any, key any, event any) error {
					// Verify the reconciliation event has correct structure
					reconEvent, ok := event.(model.ReconciliationEvent)
					assert.True(t, ok)
					assert.Equal(t, "test-task-id", reconEvent.TaskID)
					return nil
				})
			},
			expectedError: false,
		},
		{
			name: "Error downloading transaction file",
			event: model.CompilerEvent{
				Transaction:   transactionFile,
				BankStatement: bankStatementFile,
				TaskID:        "test-task-id",
				BankName:      "TestBank",
			},
			setupMocks: func(t *testing.T, m *mockFileSetup, filePath string) {
				// Mock download failure
				m.gcsRepo.EXPECT().
					DownloadFromBucket(gomock.Any(), transactionFile).
					Return(nil, errors.New("download failed"))
			},
			expectedError:  true,
			expectedErrMsg: "error starting file streamer Bank File",
		},
		{
			name: "Error downloading bank statement file",
			event: model.CompilerEvent{
				Transaction:   transactionFile,
				BankStatement: bankStatementFile,
				TaskID:        "test-task-id",
				BankName:      "TestBank",
			},
			fileContent: "id,amount,type,timestamp\ntx123,100.50,CREDIT,2023-01-15T14:30:45Z",
			setupMocks: func(t *testing.T, m *mockFileSetup, filePath string) {
				// Mock successful transaction file download
				m.gcsRepo.EXPECT().
					DownloadFromBucket(gomock.Any(), transactionFile).
					Return(createTempFileWithContent(t, filePath), nil)

				// Expect save calls for each transaction to succeed
				m.txRepo.EXPECT().Save(gomock.Any()).Return(nil).AnyTimes()

				// Mock bank statement download failure
				m.gcsRepo.EXPECT().
					DownloadFromBucket(gomock.Any(), bankStatementFile).
					Return(nil, errors.New("download failed"))
			},
			expectedError:  true,
			expectedErrMsg: "error starting file streamer Bank File",
		},
		{
			name: "Error saving transaction",
			event: model.CompilerEvent{
				Transaction:   transactionFile,
				BankStatement: bankStatementFile,
				TaskID:        "test-task-id",
				BankName:      "TestBank",
			},
			fileContent: "id,amount,type,timestamp\ntx123,100.50,CREDIT,2023-01-15T14:30:45Z",
			setupMocks: func(t *testing.T, m *mockFileSetup, filePath string) {
				// Mock GCS download
				m.gcsRepo.EXPECT().
					DownloadFromBucket(gomock.Any(), transactionFile).
					Return(createTempFileWithContent(t, filePath), nil)

				// Mock save error
				m.txRepo.EXPECT().Save(gomock.Any()).Return(errors.New("database error"))
			},
			expectedError:  true,
			expectedErrMsg: "error processing internal file",
		},
		{
			name: "Error saving bank statement",
			event: model.CompilerEvent{
				Transaction:   transactionFile,
				BankStatement: bankStatementFile,
				TaskID:        "test-task-id",
				BankName:      "TestBank",
			},
			fileContent: "id,amount,type,timestamp\ntx123,100.50,CREDIT,2023-01-15T14:30:45Z",
			setupMocks: func(t *testing.T, m *mockFileSetup, filePath string) {
				// Mock successful transaction processing
				m.gcsRepo.EXPECT().
					DownloadFromBucket(gomock.Any(), transactionFile).
					Return(createTempFileWithContent(t, filePath), nil)
				m.txRepo.EXPECT().Save(gomock.Any()).Return(nil).AnyTimes()

				// Mock bank file download
				m.gcsRepo.EXPECT().
					DownloadFromBucket(gomock.Any(), bankStatementFile).
					Return(createTempFileWithContent(t, "id,amount,date\n101,500.25,2023-01-15"), nil)

				// Mock save error for bank statement
				m.bankStmtRepo.EXPECT().Save(gomock.Any()).Return(errors.New("database issue"))
			},
			expectedError:  true,
			expectedErrMsg: "error processing internal file",
		},
		{
			name: "Kafka publish error",
			event: model.CompilerEvent{
				Transaction:   transactionFile,
				BankStatement: bankStatementFile,
				TaskID:        "test-task-id",
				BankName:      "TestBank",
			},
			fileContent: "id,amount,type,timestamp\ntx123,100.50,CREDIT,2023-01-15T14:30:45Z",
			setupMocks: func(t *testing.T, m *mockFileSetup, filePath string) {
				// Mock successful transaction processing
				m.gcsRepo.EXPECT().
					DownloadFromBucket(gomock.Any(), transactionFile).
					Return(createTempFileWithContent(t, filePath), nil)
				m.txRepo.EXPECT().Save(gomock.Any()).Return(nil).AnyTimes()

				// Mock successful bank statement processing
				m.gcsRepo.EXPECT().
					DownloadFromBucket(gomock.Any(), bankStatementFile).
					Return(createTempFileWithContent(t, "id,amount,date\n101,500.25,2023-01-15"), nil)
				m.bankStmtRepo.EXPECT().Save(gomock.Any()).Return(nil).AnyTimes()

				// Mock Kafka publish error
				m.kafkaRepo.EXPECT().
					Publish(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("kafka publish error"))
			},
			expectedError:  true,
			expectedErrMsg: "error publishing event to Kafka",
		},
		{
			name: "Invalid event - missing fields",
			event: model.CompilerEvent{
				TaskID:   "test-task-id",
				BankName: "TestBank",
				// Missing Transaction and BankStatement
			},
			setupMocks: func(t *testing.T, m *mockFileSetup, filePath string) {
				// No repo calls expected
			},
			expectedError:  true,
			expectedErrMsg: "bank statement and transaction is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a NEW mock controller for each test case
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockBankStmtRepo := repositorymock.NewMockBankStatementRepository(mockCtrl)
			mockTxRepo := repositorymock.NewMockInternalTransactionRepository(mockCtrl)
			mockGCSRepo := repositorymock.NewMockGCSRepository(mockCtrl)
			mockKafkaRepo := repositorymock.NewMockKafkaRepository(mockCtrl)

			// Create temp file with test content
			var tempFilePath string
			if tt.fileContent != "" {
				tempFilePath = filepath.Join(tempDir, "test.csv")
				err := os.WriteFile(tempFilePath, []byte(tt.fileContent), 0644)
				assert.NoError(t, err)
			}

			// Setup mocks
			mockSetup := &mockFileSetup{
				bankStmtRepo: mockBankStmtRepo,
				txRepo:       mockTxRepo,
				gcsRepo:      mockGCSRepo,
				kafkaRepo:    mockKafkaRepo,
			}

			// Pass testing.T to setupMocks for better assertions
			tt.setupMocks(t, mockSetup, tt.fileContent)

			// Create compiler with proper configuration for batch sizes
			compiler := usecase.NewFileCompiler(
				&config.Config{
					Bucket: config.BucketConfig{Name: "test-bucket"},
					App:    config.AppConfig{Compiler: config.CompilerConfig{BatchSize: 10}},
					Kafka:  config.KafkaConfig{Topic: config.TopicConfig{CompilerTopic: "test-topic"}},
				},
				mockGCSRepo,
				mockBankStmtRepo,
				mockTxRepo,
				mockKafkaRepo,
			)

			// Create event JSON
			eventBytes, err := json.Marshal(tt.event)
			assert.NoError(t, err)

			// Process the event
			err = compiler.ProcessEvent(eventBytes)

			// Check expectations
			if tt.expectedError {
				assert.Error(t, err)
				if tt.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper to create a temp file with content
func createTempFileWithContent(t *testing.T, content string) *os.File {
	tempFile, err := os.CreateTemp("", "testfile-*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if _, err := tempFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	if _, err := tempFile.Seek(0, 0); err != nil {
		t.Fatalf("Failed to seek in temp file: %v", err)
	}

	return tempFile
}

// mockFileSetup is a helper for test setup
type mockFileSetup struct {
	bankStmtRepo *repositorymock.MockBankStatementRepository
	txRepo       *repositorymock.MockInternalTransactionRepository
	gcsRepo      *repositorymock.MockGCSRepository
	kafkaRepo    *repositorymock.MockKafkaRepository
}

// TestParseTransactionRecord tests the parseTransactionRecord function
func TestParseTransactionRecord(t *testing.T) {
	tests := []struct {
		name          string
		record        []string
		expected      model.Transaction
		expectedError bool
	}{
		{
			name:   "Valid transaction record",
			record: []string{"tx123", "100.50", "CREDIT", "2023-01-15T14:30:45Z"},
			expected: model.Transaction{
				ID:              "tx123",
				Amount:          100.50,
				Type:            "CREDIT",
				TransactionTime: time.Date(2023, 1, 15, 14, 30, 45, 0, time.UTC),
			},
			expectedError: false,
		},
		{
			name:          "Record too short",
			record:        []string{"tx123", "100.50", "CREDIT"},
			expectedError: true,
		},
		{
			name:          "Invalid amount",
			record:        []string{"tx123", "invalid", "CREDIT", "2023-01-15T14:30:45Z"},
			expectedError: true,
		},
		{
			name:          "Invalid timestamp",
			record:        []string{"tx123", "100.50", "CREDIT", "invalid-time"},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transaction, err := usecase.ParseTransactionRecord(tt.record)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.ID, transaction.ID)
				assert.Equal(t, tt.expected.Amount, transaction.Amount)
				assert.Equal(t, tt.expected.Type, transaction.Type)
				assert.Equal(t, tt.expected.TransactionTime.Format(time.RFC3339),
					transaction.TransactionTime.Format(time.RFC3339))
			}
		})
	}
}

// TestSaveTransactionBatch tests the saveTransactionBatch function
func TestSaveTransactionBatch(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockTxRepo := repositorymock.NewMockInternalTransactionRepository(mockCtrl)
	mockKafkaRepo := repositorymock.NewMockKafkaRepository(mockCtrl)

	transactions := []model.Transaction{
		{ID: "tx1", Amount: 100.0, Type: "CREDIT"},
		{ID: "tx2", Amount: 200.0, Type: "DEBIT"},
	}

	t.Run("Save batch successfully", func(t *testing.T) {
		mockTxRepo.EXPECT().Save(transactions[0]).Return(nil)
		mockTxRepo.EXPECT().Save(transactions[1]).Return(nil)

		compiler := usecase.NewFileCompiler(
			&config.Config{},
			nil,
			nil,
			mockTxRepo,
			mockKafkaRepo,
		)

		err := compiler.SaveTransactionBatch(transactions)
		assert.NoError(t, err)
	})

	t.Run("Error saving transaction", func(t *testing.T) {
		mockTxRepo.EXPECT().Save(transactions[0]).Return(nil)
		mockTxRepo.EXPECT().Save(transactions[1]).Return(errors.New("database error"))

		compiler := usecase.NewFileCompiler(
			&config.Config{},
			nil,
			nil,
			mockTxRepo,
			mockKafkaRepo,
		)

		err := compiler.SaveTransactionBatch(transactions)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
	})
}

// TestSaveBankStatementBatch tests the SaveBankStatementBatch function
func TestSaveBankStatementBatch(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockBankStmtRepo := repositorymock.NewMockBankStatementRepository(mockCtrl)
	mockKafkaRepo := repositorymock.NewMockKafkaRepository(mockCtrl)

	// Updated to use string IDs instead of integers
	statements := []model.BankStatement{
		{ID: "bs-1", Amount: 100.0, Date: time.Now()},
		{ID: "bs-2", Amount: 200.0, Date: time.Now()},
	}

	t.Run("Save batch successfully", func(t *testing.T) {
		mockBankStmtRepo.EXPECT().Save(statements[0]).Return(nil)
		mockBankStmtRepo.EXPECT().Save(statements[1]).Return(nil)

		compiler := usecase.NewFileCompiler(
			&config.Config{},
			nil,
			mockBankStmtRepo,
			nil,
			mockKafkaRepo,
		)

		err := compiler.SaveBankStatementBatch(statements)
		assert.NoError(t, err)
	})

	t.Run("Error saving bank statement", func(t *testing.T) {
		mockBankStmtRepo.EXPECT().Save(statements[0]).Return(nil)
		mockBankStmtRepo.EXPECT().Save(statements[1]).Return(errors.New("database error"))

		compiler := usecase.NewFileCompiler(
			&config.Config{},
			nil,
			mockBankStmtRepo,
			nil,
			mockKafkaRepo,
		)

		err := compiler.SaveBankStatementBatch(statements)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
	})
}

// TestParseBankStatement tests the ParseBankStatement function
func TestParseBankStatement(t *testing.T) {
	tests := []struct {
		name          string
		record        []string
		expected      model.BankStatement
		expectedError bool
	}{
		{
			name:   "Valid bank statement record",
			record: []string{"bs-123", "500.25", "2023-01-15"},
			expected: model.BankStatement{
				ID:     "bs-123", // Updated to string ID
				Amount: 500.25,
				Date:   time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC),
			},
			expectedError: false,
		},
		{
			name:          "Record too short",
			record:        []string{"bs-123", "500.25"},
			expectedError: true,
		},
		{
			name:          "Invalid amount",
			record:        []string{"bs-123", "invalid", "2023-01-15"},
			expectedError: true,
		},
		{
			name:          "Invalid date",
			record:        []string{"bs-123", "500.25", "invalid-date"},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statement, err := usecase.ParseBankStatement(tt.record)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.ID, statement.ID)
				assert.Equal(t, tt.expected.Amount, statement.Amount)
				assert.Equal(t, tt.expected.Date.Format("2006-01-02"),
					statement.Date.Format("2006-01-02"))
			}
		})
	}
}
