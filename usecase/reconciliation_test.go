package usecase_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/aferryc/yars/model"
	mockrepository "github.com/aferryc/yars/repository/mocks"
	"github.com/aferryc/yars/usecase"
	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReconciliationService(t *testing.T) {
	// Setup the repository and service
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bankRepo := mockrepository.NewMockBankStatementRepository(ctrl)
	internalRepo := mockrepository.NewMockInternalTransactionRepository(ctrl)
	reconRepo := mockrepository.NewMockReconResultRepository(ctrl)
	uc := usecase.NewReconciliationUsecase(internalRepo, bankRepo, reconRepo)

	// Define time range for the test
	startTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2023, 1, 31, 23, 59, 59, 0, time.UTC)
	cTime := time.Date(2023, 1, 15, 12, 0, 0, 0, time.UTC) // Transaction time within range

	// Test data
	internalTransactions := []model.Transaction{
		{ID: "foo", Amount: 200, TransactionTime: cTime, Type: "CREDIT"},
		{ID: "bar", Amount: 200, TransactionTime: cTime, Type: "CREDIT"},
		{ID: "lorem", Amount: 400, TransactionTime: cTime, Type: "DEBIT"},
		{ID: "ipsum", Amount: 200, TransactionTime: cTime, Type: "CREDIT"},
	}

	// Updated bank statements to use string IDs instead of integers
	bankStatements := []model.BankStatement{
		{ID: "bs-1", Amount: 200, Date: cTime},
		{ID: "bs-2", Amount: 202, Date: cTime},
		{ID: "bs-3", Amount: -400, Date: cTime},
		{ID: "bs-4", Amount: 200, Date: cTime},
	}

	// Create reconciliation event
	reconEvent := model.ReconciliationEvent{
		TaskID:    "test-task-1",
		StartDate: startTime,
		EndDate:   endTime,
	}

	// Mock repository methods
	bankRepo.EXPECT().FetchAll(startTime, endTime).Return(model.BankStatementList{
		BankStatements: bankStatements,
	}, nil)
	internalRepo.EXPECT().FetchAll(startTime, endTime).Return(model.TransactionList{
		Transactions: internalTransactions,
	}, nil)

	// Expect the reconciliation summary to be stored
	reconRepo.EXPECT().StoreSummary(
		gomock.Any(),
		gomock.Any(),
		startTime,
		endTime,
	).Return(nil)

	// Perform reconciliation with event
	err := uc.ReconcileTransactions(reconEvent)
	require.NoError(t, err)
}

// TestReconciliationUsecase_ProcessEvent tests the ProcessEvent method
func TestReconciliationUsecase_ProcessEvent(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bankRepo := mockrepository.NewMockBankStatementRepository(ctrl)
	internalRepo := mockrepository.NewMockInternalTransactionRepository(ctrl)
	reconRepo := mockrepository.NewMockReconResultRepository(ctrl)
	uc := usecase.NewReconciliationUsecase(internalRepo, bankRepo, reconRepo)

	// Define time range for the test
	startTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2023, 1, 31, 23, 59, 59, 0, time.UTC)

	// Create event JSON
	event := model.ReconciliationEvent{
		TaskID:    "test-task-2",
		StartDate: startTime,
		EndDate:   endTime,
	}
	eventJSON, err := json.Marshal(event)
	require.NoError(t, err)

	// Test data
	internalTransactions := []model.Transaction{
		{ID: "tx1", Amount: 100, TransactionTime: startTime.Add(24 * time.Hour), Type: "CREDIT"},
	}

	// Updated bank statement to use string ID
	bankStatements := []model.BankStatement{
		{ID: "bs-100", Amount: 100, Date: startTime.Add(24 * time.Hour)},
	}

	// Setup expectations
	bankRepo.EXPECT().FetchAll(startTime, endTime).Return(model.BankStatementList{
		BankStatements: bankStatements,
	}, nil)
	internalRepo.EXPECT().FetchAll(startTime, endTime).Return(model.TransactionList{
		Transactions: internalTransactions,
	}, nil)

	// Expect the reconciliation summary to be stored
	reconRepo.EXPECT().StoreSummary(
		gomock.Any(),
		gomock.Any(),
		startTime,
		endTime,
	).Return(nil)

	// Test ProcessEvent
	err = uc.ProcessEvent(eventJSON)
	require.NoError(t, err)
}

// TestReconciliationUsecase_ProcessEvent_InvalidJSON tests error handling for invalid JSON
func TestReconciliationUsecase_ProcessEvent_InvalidJSON(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bankRepo := mockrepository.NewMockBankStatementRepository(ctrl)
	internalRepo := mockrepository.NewMockInternalTransactionRepository(ctrl)
	reconRepo := mockrepository.NewMockReconResultRepository(ctrl)
	uc := usecase.NewReconciliationUsecase(internalRepo, bankRepo, reconRepo)

	// Test with invalid JSON
	err := uc.ProcessEvent([]byte(`{"invalid": json`))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal event")
}

// TestReconciliationUsecase_ProcessEvent_MissingFields tests error handling for missing required fields
func TestReconciliationUsecase_ProcessEvent_MissingFields(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bankRepo := mockrepository.NewMockBankStatementRepository(ctrl)
	internalRepo := mockrepository.NewMockInternalTransactionRepository(ctrl)
	reconRepo := mockrepository.NewMockReconResultRepository(ctrl)
	uc := usecase.NewReconciliationUsecase(internalRepo, bankRepo, reconRepo)

	// Test with missing required fields
	err := uc.ProcessEvent([]byte(`{}`))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "startDate and endDate are required fields")
}

// TestReconciliationUsecase_ReconcileTransactions_RepositoryError tests error handling for repository errors
func TestReconciliationUsecase_ReconcileTransactions_RepositoryError(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bankRepo := mockrepository.NewMockBankStatementRepository(ctrl)
	internalRepo := mockrepository.NewMockInternalTransactionRepository(ctrl)
	reconRepo := mockrepository.NewMockReconResultRepository(ctrl)
	uc := usecase.NewReconciliationUsecase(internalRepo, bankRepo, reconRepo)

	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	reconEvent := model.ReconciliationEvent{
		TaskID:    "test-task-error",
		StartDate: startTime,
		EndDate:   endTime,
	}

	// Setup expectations - internal repo fails
	internalRepo.EXPECT().FetchAll(startTime, endTime).Return(model.TransactionList{}, assert.AnError)

	// Test
	err := uc.ReconcileTransactions(reconEvent)
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)

	// Setup expectations - bank repo fails
	internalRepo.EXPECT().FetchAll(startTime, endTime).Return(model.TransactionList{}, nil)
	bankRepo.EXPECT().FetchAll(startTime, endTime).Return(model.BankStatementList{}, assert.AnError)

	// Test
	err = uc.ReconcileTransactions(reconEvent)
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)

	// Setup expectations - reconciliation repo fails
	internalRepo.EXPECT().FetchAll(startTime, endTime).Return(model.TransactionList{
		Transactions: []model.Transaction{},
	}, nil)
	bankRepo.EXPECT().FetchAll(startTime, endTime).Return(model.BankStatementList{
		BankStatements: []model.BankStatement{},
	}, nil)
	reconRepo.EXPECT().StoreSummary(gomock.Any(), gomock.Any(), startTime, endTime).Return(assert.AnError)

	// Test
	err = uc.ReconcileTransactions(reconEvent)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to store summary")
}
