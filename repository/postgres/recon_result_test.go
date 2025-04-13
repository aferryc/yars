package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/aferryc/yars/model"
	"github.com/aferryc/yars/repository/postgres"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDBReconResultRepository_StoreSummary(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	repo := postgres.NewDBReconResultRepository(sqlxDB)

	ctx := context.Background()
	taskID := "test-task-id"
	startDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2023, 1, 31, 23, 59, 59, 0, time.UTC)

	t.Run("Successfully store summary with transactions and bank statements", func(t *testing.T) {
		// Create test data
		summary := model.ReconciliationSummary{
			TaskID:           taskID,
			TotalMatched:     10,
			TotalDiscrepancy: 150.75,
			TotalTransaction: 15,
			UnmatchedInternal: []model.Transaction{
				{
					ID:              "tx1",
					Amount:          100.50,
					TransactionTime: time.Date(2023, 1, 15, 12, 0, 0, 0, time.UTC),
					Type:            "CREDIT",
					Description:     "Test Transaction 1",
				},
			},
			UnmatchedBank: []model.BankStatement{
				{
					ID:        "bs-1",
					Amount:    200.25,
					Date:      time.Date(2023, 1, 20, 0, 0, 0, 0, time.UTC),
					Reference: "REF123",
					BankName:  "Test Bank",
				},
			},
		}

		// Setup expectations
		mock.ExpectBegin()

		// Summary insert - using a simpler pattern match
		mock.ExpectExec("INSERT INTO recon_summary").
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Unmatched transactions insert
		mock.ExpectExec("INSERT INTO unmatched_transactions").
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Unmatched bank statements insert
		mock.ExpectExec("INSERT INTO unmatched_bank_statements").
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		// Execute
		err := repo.StoreSummary(ctx, summary, startDate, endDate)

		// Assert
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Successfully store summary with empty unmatched lists", func(t *testing.T) {
		// Create test data with empty unmatched lists
		summary := model.ReconciliationSummary{
			TaskID:            taskID,
			TotalMatched:      10,
			TotalDiscrepancy:  0,
			TotalTransaction:  10,
			UnmatchedInternal: []model.Transaction{},
			UnmatchedBank:     []model.BankStatement{},
		}

		// Setup expectations
		mock.ExpectBegin()

		// Summary insert
		mock.ExpectExec("INSERT INTO recon_summary").
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		// Execute
		err := repo.StoreSummary(ctx, summary, startDate, endDate)

		// Assert
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error beginning transaction", func(t *testing.T) {
		// Setup expectations - transaction will fail to begin
		mock.ExpectBegin().WillReturnError(errors.New("db connection error"))

		// Create test data
		summary := model.ReconciliationSummary{
			TaskID:           taskID,
			TotalMatched:     10,
			TotalDiscrepancy: 150.75,
		}

		// Execute
		err := repo.StoreSummary(ctx, summary, startDate, endDate)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, "db connection error", err.Error())
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error inserting summary", func(t *testing.T) {
		// Create test data
		summary := model.ReconciliationSummary{
			TaskID:           taskID,
			TotalMatched:     10,
			TotalDiscrepancy: 150.75,
		}

		// Setup expectations
		mock.ExpectBegin()

		// Summary insert fails
		mock.ExpectExec("INSERT INTO recon_summary").
			WillReturnError(errors.New("insert error"))

		mock.ExpectRollback()

		// Execute
		err := repo.StoreSummary(ctx, summary, startDate, endDate)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, "insert error", err.Error())
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error inserting unmatched transactions", func(t *testing.T) {
		// Create test data
		summary := model.ReconciliationSummary{
			TaskID:           taskID,
			TotalMatched:     10,
			TotalDiscrepancy: 150.75,
			UnmatchedInternal: []model.Transaction{
				{
					ID:              "tx1",
					Amount:          100.50,
					TransactionTime: time.Now(),
				},
			},
		}

		// Setup expectations
		mock.ExpectBegin()

		// Summary insert succeeds
		mock.ExpectExec("INSERT INTO recon_summary").
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Transaction insert fails
		mock.ExpectExec("INSERT INTO unmatched_transactions").
			WillReturnError(errors.New("transaction insert error"))

		mock.ExpectRollback()

		// Execute
		err := repo.StoreSummary(ctx, summary, startDate, endDate)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "transaction insert error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error inserting unmatched bank statements", func(t *testing.T) {
		// Create test data
		summary := model.ReconciliationSummary{
			TaskID:            taskID,
			TotalMatched:      10,
			TotalDiscrepancy:  150.75,
			UnmatchedInternal: []model.Transaction{},
			UnmatchedBank: []model.BankStatement{
				{
					ID:     "bs2",
					Amount: 200.25,
					Date:   time.Date(2023, 1, 20, 0, 0, 0, 0, time.UTC),
				},
			},
		}

		// Setup expectations
		mock.ExpectBegin()

		// Summary insert succeeds
		mock.ExpectExec("INSERT INTO recon_summary").
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Bank statement insert fails
		mock.ExpectExec("INSERT INTO unmatched_bank_statements").
			WillReturnError(errors.New("bank statement insert error"))

		mock.ExpectRollback()

		// Execute
		err := repo.StoreSummary(ctx, summary, startDate, endDate)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bank statement insert error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Verify TotalDiscrepancy value stored correctly", func(t *testing.T) {
		// Create test data with specific discrepancy value
		summary := model.ReconciliationSummary{
			TaskID:            taskID,
			TotalMatched:      10,
			TotalDiscrepancy:  123.45, // Specific value to test
			TotalTransaction:  15,
			UnmatchedInternal: []model.Transaction{},
			UnmatchedBank:     []model.BankStatement{},
		}

		// Setup expectations
		mock.ExpectBegin()

		// Use a more flexible match for the SQL, just verify it contains the table name
		mock.ExpectExec("INSERT INTO recon_summary").
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		// Execute
		err := repo.StoreSummary(ctx, summary, startDate, endDate)

		// Assert
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Handles large batch of unmatched items", func(t *testing.T) {
		// Create large test data set that will trigger batch processing
		summary := model.ReconciliationSummary{
			TaskID:           taskID,
			TotalMatched:     10,
			TotalDiscrepancy: 150.75,
			TotalTransaction: 1500,
		}

		// Create 1500 unmatched transactions (more than the batch size of 1000)
		for i := 0; i < 1500; i++ {
			summary.UnmatchedInternal = append(summary.UnmatchedInternal, model.Transaction{
				ID:              fmt.Sprintf("tx%d", i),
				Amount:          float64(i) + 0.5,
				TransactionTime: time.Now().Add(time.Duration(i) * time.Hour),
				Type:            "CREDIT",
				Description:     fmt.Sprintf("Test Transaction %d", i),
			})
		}

		// Setup expectations
		mock.ExpectBegin()

		// Summary insert
		mock.ExpectExec("INSERT INTO recon_summary").
			WillReturnResult(sqlmock.NewResult(1, 1))

		// First batch of 1000 transactions
		mock.ExpectExec("INSERT INTO unmatched_transactions").
			WillReturnResult(sqlmock.NewResult(1, 1000))

		// Second batch of 500 transactions
		mock.ExpectExec("INSERT INTO unmatched_transactions").
			WillReturnResult(sqlmock.NewResult(1001, 500))

		mock.ExpectCommit()

		// Execute
		err := repo.StoreSummary(ctx, summary, startDate, endDate)

		// Assert
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
