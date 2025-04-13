package model

import "errors"

var (
	ErrTransactionNotFound    = errors.New("transaction not found")
	ErrBankStatementNotFound  = errors.New("bank statement not found")
	ErrTransactionMismatch    = errors.New("transaction mismatch")
	ErrInvalidTransactionData = errors.New("invalid transaction data")
)
