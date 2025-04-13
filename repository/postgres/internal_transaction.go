package postgres

import (
	"errors"
	"time"

	"github.com/aferryc/yars/model"
	"github.com/jmoiron/sqlx"
)

func NewDBInternalTransactionRepository(db *sqlx.DB) *DBInternalTransactionRepository {
	return &DBInternalTransactionRepository{
		db: db,
	}
}

type DBInternalTransactionRepository struct {
	db *sqlx.DB
}

type DBTransaction struct {
	ID              string    `db:"id"`
	Amount          float64   `db:"amount"`
	Type            string    `db:"type"`
	TransactionTime time.Time `db:"transaction_time"`
}

func (r *DBInternalTransactionRepository) FetchAll(start, end time.Time) (model.TransactionList, error) {
	var dbTransactions []DBTransaction

	err := r.db.Select(&dbTransactions, "SELECT id, amount, type, transaction_time FROM transactions WHERE transaction_time BETWEEN $1 AND $2", start, end)
	if err != nil {
		return model.TransactionList{}, err
	}

	// Convert DB transactions to domain model
	transactions := make([]model.Transaction, len(dbTransactions))
	for i, dbTx := range dbTransactions {
		transactions[i] = model.Transaction{
			ID:              dbTx.ID,
			Amount:          dbTx.Amount,
			Type:            dbTx.Type,
			TransactionTime: dbTx.TransactionTime,
		}
	}

	return model.TransactionList{
		Transactions: transactions,
	}, nil
}

func (r *DBInternalTransactionRepository) Save(transaction model.Transaction) error {
	dbTx := DBTransaction{
		ID:              transaction.ID,
		Amount:          transaction.Amount,
		Type:            transaction.Type,
		TransactionTime: transaction.TransactionTime,
	}

	_, err := r.db.NamedExec(
		`INSERT INTO transactions (id, amount, type, transaction_time) 
		VALUES (:id, :amount, :type, :transaction_time)
		ON CONFLICT (id, transaction_time) DO UPDATE SET
			amount = :amount, 
			type = :type`,
		dbTx,
	)

	return err
}

func (r *DBInternalTransactionRepository) FindByID(id string) (model.Transaction, error) {
	var dbTx DBTransaction

	err := r.db.Get(&dbTx, "SELECT id, amount, type, transaction_time FROM transactions WHERE id = $1", id)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return model.Transaction{}, errors.New("transaction not found")
		}
		return model.Transaction{}, err
	}

	return model.Transaction{
		ID:              dbTx.ID,
		Amount:          dbTx.Amount,
		Type:            dbTx.Type,
		TransactionTime: dbTx.TransactionTime,
	}, nil
}
