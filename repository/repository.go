package repository

import (
	"context"
	"os"
	"time"

	"github.com/aferryc/yars/model"
	"github.com/aferryc/yars/repository/postgres"
)

type BankStatementRepository interface {
	FetchAll(start, end time.Time) (model.BankStatementList, error)
	Save(statement model.BankStatement) error
	FindByID(id int) (model.BankStatement, error)
}

// InternalTransactionRepository defines the interface for internal transaction data access.
type InternalTransactionRepository interface {
	FetchAll(start, end time.Time) (model.TransactionList, error)
	Save(transaction model.Transaction) error
	FindByID(id string) (model.Transaction, error)
}

type GCSRepository interface {
	GenerateUploadURL(objectName string, contentType string, expires time.Time) (string, error)
	GenerateDownloadURL(objectName string) (string, error)
	DownloadFromBucket(ctx context.Context, objectName string) (*os.File, error)
}

type KafkaRepository interface {
	Publish(ctx context.Context, topic string, key string, message any) error
	Close() error
}

type ReconResultRepository interface {
	StoreSummary(ctx context.Context, summary model.ReconciliationSummary, startDate, endDate time.Time) error
	GetUnmatchedTransactions(ctx context.Context, taskID string, limit, offset int) ([]postgres.UnmatchedTransaction, int, error)
	GetUnmatchedBankStatements(ctx context.Context, taskID string, limit, offset int) ([]postgres.UnmatchedBankStatement, int, error)
	ListSummaries(ctx context.Context, limit, offset int) ([]postgres.ReconSummary, int, error)
}
