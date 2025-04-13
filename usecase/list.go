package usecase

import (
	"context"

	"github.com/aferryc/yars/model"
	"github.com/aferryc/yars/repository"
)

// ListUsecase handles listing operations for reconciliation data
type ListUsecase struct {
	reconRepo repository.ReconResultRepository
}

// NewListUsecase creates a new instance of ListUsecase
func NewListUsecase(reconRepo repository.ReconResultRepository) *ListUsecase {
	return &ListUsecase{
		reconRepo: reconRepo,
	}
}

// ListUnmatchedTransactions retrieves all unmatched transactions by task ID
func (u *ListUsecase) ListUnmatchedTransactions(ctx context.Context, taskID string, limit, offset int) (*model.PaginatedResponse, error) {
	// Get unmatched transactions from repository
	dbTransactions, totalCount, err := u.reconRepo.GetUnmatchedTransactions(ctx, taskID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Convert DB entities to response DTOs
	result := make([]model.UnmatchedTransactionResponse, len(dbTransactions))
	for i, tx := range dbTransactions {
		result[i] = model.UnmatchedTransactionResponse{
			ID:              tx.ID,
			TaskID:          tx.TaskID,
			Amount:          tx.Amount,
			TransactionTime: tx.TransactionTime,
			Type:            tx.Type,
			Description:     tx.Description,
		}
	}

	return &model.PaginatedResponse{
		Data:       result,
		TotalCount: totalCount,
		Limit:      limit,
		Offset:     offset,
	}, nil
}

// ListUnmatchedBankStatements retrieves all unmatched bank statements by task ID
func (u *ListUsecase) ListUnmatchedBankStatements(ctx context.Context, taskID string, limit, offset int) (*model.PaginatedResponse, error) {
	// Get unmatched bank statements from repository
	dbStatements, totalCount, err := u.reconRepo.GetUnmatchedBankStatements(ctx, taskID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Convert DB entities to response DTOs
	result := make([]model.UnmatchedBankStatementResponse, len(dbStatements))
	for i, stmt := range dbStatements {
		result[i] = model.UnmatchedBankStatementResponse{
			ID:        stmt.ID,
			TaskID:    stmt.TaskID,
			Amount:    stmt.Amount,
			Date:      stmt.Date,
			Reference: stmt.Reference,
			BankName:  stmt.BankName,
		}
	}

	return &model.PaginatedResponse{
		Data:       result,
		TotalCount: totalCount,
		Limit:      limit,
		Offset:     offset,
	}, nil
}

// ListReconSummaries retrieves a paginated list of reconciliation summaries
func (u *ListUsecase) ListReconSummaries(ctx context.Context, limit, offset int) (*model.PaginatedResponse, error) {
	// Get summaries from repository with pagination
	dbSummaries, totalCount, err := u.reconRepo.ListSummaries(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	// Convert DB entities to response DTOs
	result := make([]model.ReconSummaryResponse, len(dbSummaries))
	for i, summary := range dbSummaries {
		result[i] = model.ReconSummaryResponse{
			TaskID:                 summary.TaskID,
			TotalMatched:           summary.TotalMatched,
			TotalDiscrepancy:       summary.TotalDiscrepancy,
			TotalTransaction:       summary.TotalTransaction,
			TotalUnmatchedBank:     summary.TotalUnmatchedBank,
			TotalUnmatchedInternal: summary.TotalUnmatchedInternal,
			StartDate:              summary.StartDate,
			EndDate:                summary.EndDate,
			CreatedAt:              summary.CreatedAt,
			UpdatedAt:              summary.UpdatedAt,
		}
	}

	return &model.PaginatedResponse{
		Data:       result,
		TotalCount: totalCount,
		Limit:      limit,
		Offset:     offset,
	}, nil
}
