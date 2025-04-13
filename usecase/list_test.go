package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aferryc/yars/model"
	repositorymock "github.com/aferryc/yars/repository/mocks"
	"github.com/aferryc/yars/repository/postgres"
	"github.com/aferryc/yars/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestListUnmatchedTransactions(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositorymock.NewMockReconResultRepository(ctrl)
	useCase := usecase.NewListUsecase(mockRepo)
	ctx := context.Background()
	taskID := "test-task-id"
	limit := 10
	offset := 0

	t.Run("Successfully retrieve unmatched transactions", func(t *testing.T) {
		// Test data
		mockTransactions := []postgres.UnmatchedTransaction{
			{
				ID:              "tx1",
				TaskID:          taskID,
				Amount:          100.50,
				TransactionTime: time.Date(2023, 1, 15, 12, 0, 0, 0, time.UTC),
				Type:            "CREDIT",
				Description:     "Test Transaction 1",
			},
			{
				ID:              "tx2",
				TaskID:          taskID,
				Amount:          200.75,
				TransactionTime: time.Date(2023, 1, 16, 14, 0, 0, 0, time.UTC),
				Type:            "DEBIT",
				Description:     "Test Transaction 2",
			},
		}
		totalCount := 15 // Total transactions in database

		// Set expectations - include limit, offset, and return total count
		mockRepo.EXPECT().
			GetUnmatchedTransactions(gomock.Any(), taskID, limit, offset).
			Return(mockTransactions, totalCount, nil)

		// Execute
		result, err := useCase.ListUnmatchedTransactions(ctx, taskID, limit, offset)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, result)

		// Check pagination metadata
		assert.Equal(t, totalCount, result.TotalCount)
		assert.Equal(t, limit, result.Limit)
		assert.Equal(t, offset, result.Offset)

		// Check data content
		transactions, ok := result.Data.([]model.UnmatchedTransactionResponse)
		require.True(t, ok, "Data should be of type []model.UnmatchedTransactionResponse")
		assert.Len(t, transactions, 2)
		assert.Equal(t, "tx1", transactions[0].ID)
		assert.Equal(t, taskID, transactions[0].TaskID)
		assert.Equal(t, 100.50, transactions[0].Amount)
		assert.Equal(t, "CREDIT", transactions[0].Type)
		assert.Equal(t, "Test Transaction 1", transactions[0].Description)
	})

	t.Run("Empty result", func(t *testing.T) {
		// Set expectations - include limit, offset, and return total count
		mockRepo.EXPECT().
			GetUnmatchedTransactions(gomock.Any(), taskID, limit, offset).
			Return([]postgres.UnmatchedTransaction{}, 0, nil)

		// Execute
		result, err := useCase.ListUnmatchedTransactions(ctx, taskID, limit, offset)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, result)

		// Check pagination metadata
		assert.Equal(t, 0, result.TotalCount)
		assert.Equal(t, limit, result.Limit)
		assert.Equal(t, offset, result.Offset)

		// Check data content
		transactions, ok := result.Data.([]model.UnmatchedTransactionResponse)
		require.True(t, ok, "Data should be of type []model.UnmatchedTransactionResponse")
		assert.Empty(t, transactions)
	})

	t.Run("Repository error", func(t *testing.T) {
		// Set expectations
		expectedErr := errors.New("database error")
		mockRepo.EXPECT().
			GetUnmatchedTransactions(gomock.Any(), taskID, limit, offset).
			Return(nil, 0, expectedErr)

		// Execute
		result, err := useCase.ListUnmatchedTransactions(ctx, taskID, limit, offset)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, result)
	})
}

func TestListUnmatchedBankStatements(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositorymock.NewMockReconResultRepository(ctrl)
	useCase := usecase.NewListUsecase(mockRepo)
	ctx := context.Background()
	taskID := "test-task-id"
	limit := 10
	offset := 0

	t.Run("Successfully retrieve unmatched bank statements", func(t *testing.T) {
		// Test data
		mockStatements := []postgres.UnmatchedBankStatement{
			{
				ID:        1,
				TaskID:    taskID,
				Amount:    100.50,
				Date:      time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC),
				Reference: "REF123",
				BankName:  "Test Bank",
			},
			{
				ID:        2,
				TaskID:    taskID,
				Amount:    200.75,
				Date:      time.Date(2023, 1, 16, 0, 0, 0, 0, time.UTC),
				Reference: "REF456",
				BankName:  "Test Bank",
			},
		}
		totalCount := 8 // Total bank statements in database

		// Set expectations - include limit, offset, and return total count
		mockRepo.EXPECT().
			GetUnmatchedBankStatements(gomock.Any(), taskID, limit, offset).
			Return(mockStatements, totalCount, nil)

		// Execute
		result, err := useCase.ListUnmatchedBankStatements(ctx, taskID, limit, offset)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, result)

		// Check pagination metadata
		assert.Equal(t, totalCount, result.TotalCount)
		assert.Equal(t, limit, result.Limit)
		assert.Equal(t, offset, result.Offset)

		// Check data content
		statements, ok := result.Data.([]model.UnmatchedBankStatementResponse)
		require.True(t, ok, "Data should be of type []model.UnmatchedBankStatementResponse")
		assert.Len(t, statements, 2)
		assert.Equal(t, 1, statements[0].ID)
		assert.Equal(t, taskID, statements[0].TaskID)
		assert.Equal(t, 100.50, statements[0].Amount)
		assert.Equal(t, "REF123", statements[0].Reference)
		assert.Equal(t, "Test Bank", statements[0].BankName)
	})

	t.Run("Empty result", func(t *testing.T) {
		// Set expectations - include limit, offset, and return total count
		mockRepo.EXPECT().
			GetUnmatchedBankStatements(gomock.Any(), taskID, limit, offset).
			Return([]postgres.UnmatchedBankStatement{}, 0, nil)

		// Execute
		result, err := useCase.ListUnmatchedBankStatements(ctx, taskID, limit, offset)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, result)

		// Check pagination metadata
		assert.Equal(t, 0, result.TotalCount)
		assert.Equal(t, limit, result.Limit)
		assert.Equal(t, offset, result.Offset)

		// Check data content
		statements, ok := result.Data.([]model.UnmatchedBankStatementResponse)
		require.True(t, ok, "Data should be of type []model.UnmatchedBankStatementResponse")
		assert.Empty(t, statements)
	})

	t.Run("Repository error", func(t *testing.T) {
		// Set expectations
		expectedErr := errors.New("database error")
		mockRepo.EXPECT().
			GetUnmatchedBankStatements(gomock.Any(), taskID, limit, offset).
			Return(nil, 0, expectedErr)

		// Execute
		result, err := useCase.ListUnmatchedBankStatements(ctx, taskID, limit, offset)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, result)
	})
}

func TestListReconSummaries(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositorymock.NewMockReconResultRepository(ctrl)
	useCase := usecase.NewListUsecase(mockRepo)
	ctx := context.Background()
	limit := 10
	offset := 0

	t.Run("Successfully retrieve summaries", func(t *testing.T) {
		// Test data
		startDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2023, 1, 31, 23, 59, 59, 0, time.UTC)
		createdAt := time.Date(2023, 2, 1, 10, 0, 0, 0, time.UTC)
		updatedAt := time.Date(2023, 2, 1, 10, 0, 0, 0, time.UTC)

		mockSummaries := []postgres.ReconSummary{
			{
				TaskID:                 "task1",
				TotalMatched:           10,
				TotalDiscrepancy:       150.75,
				TotalTransaction:       15,
				TotalUnmatchedBank:     3,
				TotalUnmatchedInternal: 2,
				StartDate:              startDate,
				EndDate:                endDate,
				CreatedAt:              createdAt,
				UpdatedAt:              updatedAt,
			},
			{
				TaskID:                 "task2",
				TotalMatched:           20,
				TotalDiscrepancy:       75.25,
				TotalTransaction:       25,
				TotalUnmatchedBank:     2,
				TotalUnmatchedInternal: 3,
				StartDate:              startDate,
				EndDate:                endDate,
				CreatedAt:              createdAt,
				UpdatedAt:              updatedAt,
			},
		}
		totalCount := 25 // Total summaries in database

		// Set expectations - return total count
		mockRepo.EXPECT().
			ListSummaries(gomock.Any(), limit, offset).
			Return(mockSummaries, totalCount, nil)

		// Execute
		result, err := useCase.ListReconSummaries(ctx, limit, offset)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, result)

		// Check pagination metadata
		assert.Equal(t, totalCount, result.TotalCount)
		assert.Equal(t, limit, result.Limit)
		assert.Equal(t, offset, result.Offset)

		// Check data content
		summaries, ok := result.Data.([]model.ReconSummaryResponse)
		require.True(t, ok, "Data should be of type []model.ReconSummaryResponse")
		assert.Len(t, summaries, 2)
		assert.Equal(t, "task1", summaries[0].TaskID)
		assert.Equal(t, 10, summaries[0].TotalMatched)
		assert.Equal(t, 150.75, summaries[0].TotalDiscrepancy)
		assert.Equal(t, 15, summaries[0].TotalTransaction)
		assert.Equal(t, 3, summaries[0].TotalUnmatchedBank)
		assert.Equal(t, 2, summaries[0].TotalUnmatchedInternal)
		assert.Equal(t, startDate, summaries[0].StartDate)
		assert.Equal(t, endDate, summaries[0].EndDate)
		assert.Equal(t, createdAt, summaries[0].CreatedAt)
		assert.Equal(t, updatedAt, summaries[0].UpdatedAt)
	})

	t.Run("Empty result", func(t *testing.T) {
		// Set expectations - return total count
		mockRepo.EXPECT().
			ListSummaries(gomock.Any(), limit, offset).
			Return([]postgres.ReconSummary{}, 0, nil)

		// Execute
		result, err := useCase.ListReconSummaries(ctx, limit, offset)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, result)

		// Check pagination metadata
		assert.Equal(t, 0, result.TotalCount)
		assert.Equal(t, limit, result.Limit)
		assert.Equal(t, offset, result.Offset)

		// Check data content
		summaries, ok := result.Data.([]model.ReconSummaryResponse)
		require.True(t, ok, "Data should be of type []model.ReconSummaryResponse")
		assert.Empty(t, summaries)
	})

	t.Run("Repository error", func(t *testing.T) {
		// Set expectations
		expectedErr := errors.New("database error")
		mockRepo.EXPECT().
			ListSummaries(gomock.Any(), limit, offset).
			Return(nil, 0, expectedErr)

		// Execute
		result, err := useCase.ListReconSummaries(ctx, limit, offset)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, result)
	})
}
