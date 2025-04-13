package postgres

import (
	"errors"
	"time"

	"github.com/aferryc/yars/model"
	"github.com/jmoiron/sqlx"
)

type DBBankStatementRepository struct {
	db *sqlx.DB
}

type DBBankStatement struct {
	ID     string    `db:"id"`
	Amount float64   `db:"amount"`
	Date   time.Time `db:"date"`
	Bank   string    `db:"bank"`
}

func NewDBBankStatementRepository(db *sqlx.DB) *DBBankStatementRepository {
	return &DBBankStatementRepository{
		db: db,
	}
}

func (r *DBBankStatementRepository) FetchAll(start, end time.Time) (model.BankStatementList, error) {
	var dbStatements []DBBankStatement

	err := r.db.Select(&dbStatements, "SELECT id, amount, date, bank FROM bank_statements WHERE date BETWEEN $1 AND $2", start, end)
	if err != nil {
		return model.BankStatementList{}, err
	}

	statements := make([]model.BankStatement, len(dbStatements))
	for i, dbStmt := range dbStatements {
		statements[i] = model.BankStatement{
			ID:     dbStmt.ID,
			Amount: dbStmt.Amount,
			Date:   dbStmt.Date,
		}
	}

	return model.BankStatementList{
		BankStatements: statements,
	}, nil
}

func (r *DBBankStatementRepository) Save(statement model.BankStatement) error {
	dbStmt := DBBankStatement{
		ID:     statement.ID,
		Amount: statement.Amount,
		Date:   statement.Date,
	}

	query := `
	INSERT INTO bank_statements (id, amount, date, bank) 
	VALUES (:id, :amount, :date, :bank)
	ON CONFLICT (id, date, bank) DO UPDATE SET
		amount = :amount
	`

	_, err := r.db.NamedExec(query, dbStmt)
	return err
}

func (r *DBBankStatementRepository) FindByID(id int) (model.BankStatement, error) {
	var dbStmt DBBankStatement

	err := r.db.Get(&dbStmt, "SELECT id, amount, date, bank FROM bank_statements WHERE id = $1", id)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return model.BankStatement{}, errors.New("bank statement not found")
		}
		return model.BankStatement{}, err
	}

	return model.BankStatement{
		ID:     dbStmt.ID,
		Amount: dbStmt.Amount,
		Date:   dbStmt.Date,
	}, nil
}
