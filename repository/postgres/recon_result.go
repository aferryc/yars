package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/aferryc/yars/internal/utils"
	"github.com/aferryc/yars/model"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func NewDBReconResultRepository(db *sqlx.DB) *DBReconResultRepository {
	return &DBReconResultRepository{
		db: db,
	}
}

type DBReconResultRepository struct {
	db *sqlx.DB
}

type ReconSummary struct {
	TaskID                 string    `db:"id"`
	TotalMatched           int       `db:"matched"`
	TotalDiscrepancy       float64   `db:"discrepancy"`
	TotalTransaction       int       `db:"total_transaction"`
	TotalUnmatchedBank     int       `db:"total_unmatched_bank"`
	TotalUnmatchedInternal int       `db:"total_unmatched_internal"`
	StartDate              time.Time `db:"start_date"`
	EndDate                time.Time `db:"end_date"`
	CreatedAt              time.Time `db:"created_at"`
	UpdatedAt              time.Time `db:"updated_at"`
}

type UnmatchedTransaction struct {
	ID              string    `db:"id"`
	TaskID          string    `db:"task_id"`
	Amount          float64   `db:"amount"`
	TransactionTime time.Time `db:"transaction_time"`
	Type            string    `db:"type"`
	Description     string    `db:"description"`
	CreatedAt       time.Time `db:"created_at"`
}

type UnmatchedBankStatement struct {
	ID        int       `db:"id"`
	TaskID    string    `db:"task_id"`
	Amount    float64   `db:"amount"`
	Date      time.Time `db:"date"`
	Reference string    `db:"reference"`
	BankName  string    `db:"bank_name"`
	CreatedAt time.Time `db:"created_at"`
}

const batchSize = 1000

func (r *DBReconResultRepository) StoreSummary(ctx context.Context, summary model.ReconciliationSummary, startDate, endDate time.Time) error {
	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = r.insertReconSummary(ctx, tx, summary, startDate, endDate); err != nil {
		return err
	}

	if err = r.insertUnmatchedTransactions(ctx, tx, summary.TaskID, summary.UnmatchedInternal); err != nil {
		return err
	}

	if err = r.insertUnmatchedBankStatements(ctx, tx, summary.TaskID, summary.UnmatchedBank); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *DBReconResultRepository) insertReconSummary(ctx context.Context, tx *sqlx.Tx, summary model.ReconciliationSummary, startDate, endDate time.Time) error {
	_, err := tx.NamedExecContext(ctx, `
		INSERT INTO recon_summary (
			id, matched, discrepancy, total_transaction, 
			total_unmatched_bank, total_unmatched_internal,
			start_date, end_date, created_at, updated_at
		) VALUES (
			:id, :matched, :discrepancy, :total_transaction,
			:total_unmatched_bank, :total_unmatched_internal,
			:start_date, :end_date, NOW(), NOW()
		)`,
		ReconSummary{
			TaskID:                 summary.TaskID,
			TotalMatched:           summary.TotalMatched,
			TotalDiscrepancy:       summary.TotalDiscrepancy,
			TotalTransaction:       summary.TotalTransaction,
			TotalUnmatchedBank:     len(summary.UnmatchedBank),
			TotalUnmatchedInternal: len(summary.UnmatchedInternal),
			StartDate:              startDate,
			EndDate:                endDate,
		})
	return err
}

func (r *DBReconResultRepository) insertUnmatchedTransactions(ctx context.Context, tx *sqlx.Tx, taskID string, unmatchedTxns []model.Transaction) error {
	if len(unmatchedTxns) == 0 {
		return nil
	}

	unmatchedTxnsChunks := utils.ChunkSlice(unmatchedTxns, batchSize)
	for _, chunk := range unmatchedTxnsChunks {
		if err := r.insertUnmatchedTransactionsBatch(ctx, tx, taskID, chunk); err != nil {
			return errors.Wrap(err, "[insertUnmatchedTransactions] error inserting unmatched transactions batch")
		}
	}
	return nil
}

func (r *DBReconResultRepository) insertUnmatchedTransactionsBatch(ctx context.Context, tx *sqlx.Tx, taskID string, unmatchedTxns []model.Transaction) error {
	query := `
		INSERT INTO unmatched_transactions (
			id, task_id, amount, transaction_time, type, description
		) VALUES (
			:id, :task_id, :amount, :transaction_time, :type, :description
		)`

	records := make([]UnmatchedTransaction, len(unmatchedTxns))
	for j, txn := range unmatchedTxns {
		records[j] = UnmatchedTransaction{
			ID:              txn.ID,
			TaskID:          taskID,
			Amount:          txn.Amount,
			TransactionTime: txn.TransactionTime,
			Type:            txn.Type,
			Description:     txn.Description,
		}
	}

	_, err := tx.NamedExecContext(ctx, query, records)
	if err != nil {
		return err
	}

	return nil
}

func (r *DBReconResultRepository) insertUnmatchedBankStatements(ctx context.Context, tx *sqlx.Tx, taskID string, unmatchedStmts []model.BankStatement) error {
	if len(unmatchedStmts) == 0 {
		return nil
	}

	unmatchedStmtsChunks := utils.ChunkSlice(unmatchedStmts, batchSize)

	for _, chunk := range unmatchedStmtsChunks {
		if err := r.insertUnmatchedBankStatementsBatch(ctx, tx, taskID, chunk); err != nil {
			return errors.Wrap(err, "[insertUnmatchedBankStatements] error inserting unmatched bank statements batch")
		}
	}
	return nil
}

func (r *DBReconResultRepository) insertUnmatchedBankStatementsBatch(ctx context.Context, tx *sqlx.Tx, taskID string, unmatchedStmts []model.BankStatement) error {
	query := `
		INSERT INTO unmatched_bank_statements (
			task_id, amount, date, reference, bank_name
		) VALUES (
			:task_id, :amount, :date, :reference, :bank_name
		)`

	records := make([]UnmatchedBankStatement, len(unmatchedStmts))
	for j, stmt := range unmatchedStmts {
		records[j] = UnmatchedBankStatement{
			TaskID:    taskID,
			Amount:    stmt.Amount,
			Date:      stmt.Date,
			Reference: stmt.Reference,
			BankName:  stmt.BankName,
		}
	}

	_, err := tx.NamedExecContext(ctx, query, records)
	if err != nil {
		return errors.Wrap(err, "[insertUnmatchedBankStatementsBatch] error inserting unmatched bank statements")
	}

	return nil
}

func (r *DBReconResultRepository) GetUnmatchedTransactions(ctx context.Context, taskID string, limit, offset int) ([]UnmatchedTransaction, int, error) {
	var transactions []UnmatchedTransaction
	err := r.db.SelectContext(ctx, &transactions, `
		SELECT * FROM unmatched_transactions 
		WHERE task_id = $1
		ORDER BY transaction_time DESC
		LIMIT $2 OFFSET $3`, taskID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	var total int
	err = r.db.GetContext(ctx, &total, `
		SELECT COUNT(*) FROM unmatched_bank_statements 
		WHERE task_id = $1`, taskID)
	if err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

func (r *DBReconResultRepository) GetUnmatchedBankStatements(ctx context.Context, taskID string, limit, offset int) ([]UnmatchedBankStatement, int, error) {
	var statements []UnmatchedBankStatement
	err := r.db.SelectContext(ctx, &statements, `
		SELECT * FROM unmatched_bank_statements 
		WHERE task_id = $1
		ORDER BY date DESC
		LIMIT $2 OFFSET $3`, taskID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	var total int
	err = r.db.GetContext(ctx, &total, `
		SELECT COUNT(*) FROM unmatched_bank_statements 
		WHERE task_id = $1`, taskID)
	if err != nil {
		return nil, 0, err
	}

	return statements, total, nil
}

func (r *DBReconResultRepository) ListSummaries(ctx context.Context, limit, offset int) ([]ReconSummary, int, error) {
	var summaries []ReconSummary
	err := r.db.SelectContext(ctx, &summaries, `
		SELECT * FROM recon_summary 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// Get total count for pagination
	var total int
	err = r.db.GetContext(ctx, &total, `
		SELECT COUNT(*) FROM recon_summary`)
	if err != nil {
		return nil, 0, err
	}

	return summaries, total, nil
}
