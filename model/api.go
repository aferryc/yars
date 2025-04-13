package model

import "time"

type UploadURLResponse struct {
	TransactionURL   string    `json:"transactionUrl"`
	BankStatementURL string    `json:"bankStatementUrl"`
	TaskID           string    `json:"taskID"`
	ExpiresAt        time.Time `json:"expiresAt"`
}

type CompilerRequest struct {
	TaskID    string    `json:"taskID"`
	BankName  string    `json:"bankName"`
	StartDate time.Time `json:"startDate,omitempty"`
	EndDate   time.Time `json:"endDate,omitempty"`
}

type ReconSummaryResponse struct {
	TaskID                 string    `json:"taskId"`
	TotalMatched           int       `json:"totalMatched"`
	TotalDiscrepancy       float64   `json:"totalDiscrepancy"`
	TotalTransaction       int       `json:"totalTransaction"`
	TotalUnmatchedBank     int       `json:"totalUnmatchedBank"`
	TotalUnmatchedInternal int       `json:"totalUnmatchedInternal"`
	StartDate              time.Time `json:"startDate"`
	EndDate                time.Time `json:"endDate"`
	CreatedAt              time.Time `json:"createdAt"`
	UpdatedAt              time.Time `json:"updatedAt"`
}

type UnmatchedTransactionResponse struct {
	ID              string    `json:"id"`
	TaskID          string    `json:"taskId"`
	Amount          float64   `json:"amount"`
	TransactionTime time.Time `json:"transactionTime"`
	Type            string    `json:"type"`
	Description     string    `json:"description"`
}

type UnmatchedBankStatementResponse struct {
	ID        int       `json:"id"`
	TaskID    string    `json:"taskId"`
	Amount    float64   `json:"amount"`
	Date      time.Time `json:"date"`
	Reference string    `json:"reference"`
	BankName  string    `json:"bankName"`
}

type PaginatedResponse struct {
	Data       any `json:"data"`
	TotalCount int `json:"totalCount"`
	Limit      int `json:"limit"`
	Offset     int `json:"offset"`
}
