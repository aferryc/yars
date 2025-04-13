package model

import (
	"fmt"
	"time"
)

const amountFormat = "%.2f"
const TransactionFile = "transactions.csv"
const BankStatementFile = "bank_statement.csv"

type Transaction struct {
	ID              string    `json:"id"`
	Amount          float64   `json:"amount"`
	TransactionTime time.Time `json:"date"`
	Type            string    `json:"type"` // e.g., "DEBIT" or "CREDIT"
	Description     string    `json:"description"`
}

type TransactionList struct {
	Transactions []Transaction `json:"transactions"`
}

type TransactionPrecompiled struct {
	Count map[string]int
	List  map[string][]Transaction
}

func (t TransactionList) Precompile() TransactionPrecompiled {
	countlist := make(map[string]int)
	list := make(map[string][]Transaction)
	for _, tx := range t.Transactions {
		amount := tx.Amount
		if tx.Type == "DEBIT" {
			amount = -amount
		}

		// Normalize the amount to a string with two decimal places
		// This is important for consistent string keys in the map
		// and to avoid floating-point precision issues
		amountStr := fmt.Sprintf(amountFormat, amount)
		countlist[amountStr]++
		list[amountStr] = append(list[amountStr], tx)
	}

	return TransactionPrecompiled{
		Count: countlist,
		List:  list,
	}
}

type BankStatement struct {
	ID        string    `json:"id"`
	Amount    float64   `json:"amount"`
	Date      time.Time `json:"date"`
	Reference string    `json:"reference"`
	BankName  string    `json:"bank_name"`
}

type BankStatementList struct {
	BankStatements []BankStatement `json:"bank_statements"`
}

type BankStatementPrecompiled struct {
	Count map[string]int
	List  map[string][]BankStatement
}

func (t BankStatementList) Precompile() BankStatementPrecompiled {
	countlist := make(map[string]int)
	list := make(map[string][]BankStatement)
	for _, tx := range t.BankStatements {
		amountStr := fmt.Sprintf(amountFormat, tx.Amount)
		countlist[amountStr]++
		list[amountStr] = append(list[amountStr], tx)
	}

	return BankStatementPrecompiled{
		Count: countlist,
		List:  list,
	}
}

type ReconciliationSummary struct {
	UnmatchedInternal []Transaction
	UnmatchedBank     []BankStatement
	TotalMatched      int
	TotalDiscrepancy  float64
	TotalTransaction  int
	TaskID            string
}
