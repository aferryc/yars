package model

import "time"

type CompilerEvent struct {
	BankStatement string    `json:"bank_statement"`
	Transaction   string    `json:"transaction"`
	BankName      string    `json:"bankName"`
	StartDate     time.Time `json:"startDate,omitempty"`
	EndDate       time.Time `json:"endDate,omitempty"`
	TaskID        string    `json:"taskID"`
}

type ReconciliationEvent struct {
	TaskID    string    `json:"taskID"`
	StartDate time.Time `json:"startDate,omitempty"`
	EndDate   time.Time `json:"endDate,omitempty"`
}
