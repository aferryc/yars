package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aferryc/yars/model"
	"github.com/aferryc/yars/repository"
	"github.com/pkg/errors"
)

type ReconciliationUsecase struct {
	internalRepo repository.InternalTransactionRepository
	bankRepo     repository.BankStatementRepository
	reconRepo    repository.ReconResultRepository
}

func NewReconciliationUsecase(internalRepo repository.InternalTransactionRepository, bankRepo repository.BankStatementRepository, reconRepo repository.ReconResultRepository) *ReconciliationUsecase {
	return &ReconciliationUsecase{
		internalRepo: internalRepo,
		bankRepo:     bankRepo,
		reconRepo:    reconRepo,
	}
}

func (r *ReconciliationUsecase) ProcessEvent(event []byte) error {
	var reconEvent model.ReconciliationEvent
	err := json.Unmarshal(event, &reconEvent)
	if err != nil {
		return errors.Wrap(err, "[parseEvent] failed to unmarshal event")
	}
	if reconEvent.StartDate.IsZero() || reconEvent.EndDate.IsZero() {
		return errors.Wrap(errors.New("startDate and endDate are required fields"), "[parseEvent] invalid event")
	}

	return r.ReconcileTransactions(reconEvent)
}

func (r *ReconciliationUsecase) ReconcileTransactions(event model.ReconciliationEvent) error {
	ctx := context.Background()
	internalTransactions, err := r.internalRepo.FetchAll(event.StartDate, event.EndDate)
	if err != nil {
		return err
	}

	bankStatements, err := r.bankRepo.FetchAll(event.StartDate, event.EndDate)
	if err != nil {
		return err
	}

	summary := r.matchTransactions(internalTransactions, bankStatements)
	summary.TaskID = event.TaskID

	err = r.reconRepo.StoreSummary(ctx, summary, event.StartDate, event.EndDate)
	if err != nil {
		return errors.Wrap(err, "[ReconcileTransactions] failed to store summary")

	}

	return nil
}

func (r *ReconciliationUsecase) matchTransactions(internalTransactions model.TransactionList, bankStatements model.BankStatementList) model.ReconciliationSummary {
	var matchedCount int
	tx := internalTransactions.Precompile()
	bank := bankStatements.Precompile()
	var unmatchedInternal []model.Transaction
	var unmatchedBank []model.BankStatement

	matchedCount = 0

	for amount, count := range tx.Count {
		bankCount := bank.Count[amount]

		matchesAtThisAmount := min(count, bankCount)
		matchedCount += matchesAtThisAmount

		if count > bankCount {
			unmatchedCount := count - bankCount

			// Take the last unmatchedCount transactions from this amount group
			transactions := tx.List[amount]
			startIndex := len(transactions) - unmatchedCount
			unmatchedInternal = append(unmatchedInternal, transactions[startIndex:]...)
		}
	}

	for amount, count := range bank.Count {
		internalCount := tx.Count[amount]

		if count > internalCount {
			unmatchedCount := count - internalCount
			// Take the last unmatchedCount statements from this amount group
			statements := bank.List[amount]
			startIndex := len(statements) - unmatchedCount
			unmatchedBank = append(unmatchedBank, statements[startIndex:]...)
		}
	}

	totalDiscrepancy := sumTotalDiscrepancy(unmatchedBank, unmatchedInternal)

	return model.ReconciliationSummary{
		UnmatchedInternal: unmatchedInternal,
		UnmatchedBank:     unmatchedBank,
		TotalMatched:      matchedCount,
		TotalDiscrepancy:  totalDiscrepancy,
		TotalTransaction:  matchedCount + len(unmatchedInternal) + len(unmatchedBank),
	}
}

func sumTotalDiscrepancy(unmatchedBank []model.BankStatement, unmatchedInternal []model.Transaction) float64 {
	var totalDiscrepancy float64
	for _, statement := range unmatchedBank {
		totalDiscrepancy += statement.Amount
	}
	for _, transaction := range unmatchedInternal {
		totalDiscrepancy += transaction.Amount
	}
	return totalDiscrepancy
}

func (r *ReconciliationUsecase) GenerateReport(unmatchedInternal []model.Transaction, unmatchedBank []model.BankStatement) string {
	report := fmt.Sprintf("Reconciliation Report - %s\n", time.Now().Format("2006-01-02 15:04:05"))
	report += fmt.Sprintf("Unmatched Internal Transactions: %d\n", len(unmatchedInternal))
	report += fmt.Sprintf("Unmatched Bank Statements: %d\n", len(unmatchedBank))
	return report
}
