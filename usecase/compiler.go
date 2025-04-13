package usecase

import (
	"context"
	"encoding/csv"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"encoding/json"

	"github.com/aferryc/yars/internal/config"
	"github.com/aferryc/yars/model"
	"github.com/aferryc/yars/repository"
)

// FileCompiler implements the CompilerUseCase
type FileCompiler struct {
	bucketName      string
	storageRepo     repository.GCSRepository
	bankStmtRepo    repository.BankStatementRepository
	transactionRepo repository.InternalTransactionRepository
	kafkaRepo       repository.KafkaRepository
	batchSize       int
	cfg             *config.Config
}

// NewFileCompiler creates a new FileCompiler instance
func NewFileCompiler(
	cfg *config.Config,
	storageRepo repository.GCSRepository,
	bankStmtRepo repository.BankStatementRepository,
	transactionRepo repository.InternalTransactionRepository,
	kafkaRepo repository.KafkaRepository,
) *FileCompiler {
	return &FileCompiler{
		cfg:             cfg,
		storageRepo:     storageRepo,
		bankStmtRepo:    bankStmtRepo,
		transactionRepo: transactionRepo,
		kafkaRepo:       kafkaRepo,
	}
}

// ProcessFile handles the entire process of downloading, parsing and storing file data
func (fc *FileCompiler) ProcessEvent(event []byte) error {
	ctx := context.Background()
	compilerEvent, err := parseEvent(event)
	if err != nil {
		return errors.Wrap(err, "[Compiler.ProcessFile] failed to unmarshal event")
	}

	for _, objectName := range []string{compilerEvent.Transaction, compilerEvent.BankStatement} {
		if err := fc.processFile(ctx, objectName, compilerEvent.BankName); err != nil {
			return errors.Wrap(err, "[Compiler.ProcessFile] error processing file")
		}
	}

	err = fc.kafkaRepo.Publish(ctx, fc.cfg.Kafka.Topic.CompilerTopic, compilerEvent.TaskID, model.ReconciliationEvent{
		TaskID:    compilerEvent.TaskID,
		StartDate: compilerEvent.StartDate,
		EndDate:   compilerEvent.EndDate,
	})
	if err != nil {
		return errors.Wrap(err, "[Compiler.ProcessFile] error publishing event to Kafka")
	}

	return nil
}

func (fc *FileCompiler) processFile(ctx context.Context, objectName, bankName string) error {
	// Check if the objectName is empty
	// This probably because user only upload other file
	if objectName == "" {
		return nil
	}

	csvReader, tempFile, err := fc.startFileStreamer(ctx, fc.bucketName, objectName)
	if err != nil {
		return errors.Wrap(err, "[Compiler.ProcessFile] error starting file streamer Bank File")
	}
	if strings.Contains(objectName, model.BankStatementFile) {
		err = fc.processBankStatement(csvReader, bankName)
	} else {
		err = fc.processInternalTransactions(csvReader)
	}
	if err != nil {
		return errors.Wrapf(err, "[Compiler.ProcessFile] error processing internal file %s", objectName)
	}
	defer func() {
		if tempFile == nil {
			return
		}
		if err := tempFile.Close(); err != nil {
			log.Printf("Error closing file: %v", err)
		}
	}()

	return nil
}

func (fc *FileCompiler) startFileStreamer(ctx context.Context, bucketName string, objectName string) (*csv.Reader, *os.File, error) {
	tempFile, err := fc.storageRepo.DownloadFromBucket(ctx, objectName)
	if err != nil {
		return nil, nil, errors.Wrap(err, "[startFileStreamer] error downloading file from GCS")
	}

	if _, err := tempFile.Seek(0, 0); err != nil {
		return nil, tempFile, errors.Wrap(err, "[startFileStreamer] error rewinding temp file")
	}

	csvReader := csv.NewReader(tempFile)

	// Skip header
	// Assuming the first line is a header
	if _, err := csvReader.Read(); err != nil {
		return csvReader, tempFile, errors.Wrap(err, "[startFileStreamer] error reading header")
	}

	return csvReader, tempFile, nil
}

func parseEvent(event []byte) (model.CompilerEvent, error) {
	var compilerEvent model.CompilerEvent
	err := json.Unmarshal(event, &compilerEvent)
	if err != nil {
		return model.CompilerEvent{}, errors.Wrap(err, "[parseEvent] failed to unmarshal event")
	}
	if compilerEvent.BankStatement == "" && compilerEvent.Transaction == "" {
		return model.CompilerEvent{}, errors.New("[parseEvent] bank statement and transaction is empty")
	}
	return compilerEvent, nil
}

func (fc *FileCompiler) processInternalTransactions(csvReader *csv.Reader) error {
	var processedCount int
	var batchSize int = 0
	var batch []model.Transaction

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading CSV record: %v", err)
			continue
		}

		transaction, err := ParseTransactionRecord(record)
		if err != nil {
			log.Printf("Error parsing transaction record: %v", err)
			continue
		}

		// Add to batch
		batch = append(batch, transaction)
		batchSize++

		if batchSize >= fc.cfg.App.Compiler.BatchSize {
			if err := fc.SaveTransactionBatch(batch); err != nil {
				return errors.Wrap(err, "[processInternalTransactions] error saving transaction during batch")
			}
			processedCount += batchSize
			batchSize = 0
			batch = nil
		}
	}

	if batchSize > 0 {
		if err := fc.SaveTransactionBatch(batch); err != nil {
			return errors.Wrap(err, "[processInternalTransactions] error saving transaction batch")
		}
		processedCount += batchSize
	}

	log.Printf("Processed %d internal transactions", processedCount)
	return nil
}

func ParseTransactionRecord(record []string) (model.Transaction, error) {
	if len(record) < 4 {
		return model.Transaction{}, errors.Wrap(errors.New("invalid record format"), "[parseTransactionRecord] error parsing transaction record")
	}

	amount, err := strconv.ParseFloat(record[1], 64)
	if err != nil {
		return model.Transaction{}, errors.Wrap(err, "[parseTransactionRecord] error parsing amount")
	}

	txTime, err := time.Parse("2006-01-02T15:04:05Z", record[3])
	if err != nil {
		return model.Transaction{}, errors.Wrap(err, "[parseTransactionRecord] error parsing transaction time")
	}
	return model.Transaction{
		ID:              record[0],
		Amount:          amount,
		TransactionTime: txTime,
		Type:            record[2],
	}, nil
}

// saveTransactionBatch saves a batch of transactions to the database
func (fc *FileCompiler) SaveTransactionBatch(transactions []model.Transaction) error {
	for _, tx := range transactions {
		if err := fc.transactionRepo.Save(tx); err != nil {
			return err
		}
	}
	return nil
}

func (fc *FileCompiler) processBankStatement(csvReader *csv.Reader, bankName string) error {
	var processedCount int
	var batchSize int = 0
	var batch []model.BankStatement

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading CSV record: %v", err)
			continue
		}

		stmt, err := ParseBankStatement(record)
		if err != nil {
			log.Printf("Error parsing bank statement record: %v", err)
			continue
		}

		batch = append(batch, stmt)
		batchSize++

		if batchSize >= 100 {
			if err := fc.SaveBankStatementBatch(batch); err != nil {
				return errors.Wrap(err, "[processBankStatments] error saving transaction inside batch")
			}
			processedCount += batchSize
			batchSize = 0
			batch = nil
		}
	}

	if batchSize > 0 {
		if err := fc.SaveBankStatementBatch(batch); err != nil {
			return errors.Wrap(err, "[processBankStatments] error saving transaction batch")
		}
		processedCount += batchSize
	}

	log.Printf("Processed %d bank statements for %s", processedCount, bankName)
	return nil
}

func (fc *FileCompiler) SaveBankStatementBatch(statements []model.BankStatement) error {
	for _, stmt := range statements {
		if err := fc.bankStmtRepo.Save(stmt); err != nil {
			return err
		}
	}
	return nil
}

func ParseBankStatement(record []string) (model.BankStatement, error) {
	if len(record) < 3 {
		return model.BankStatement{}, errors.Wrap(errors.New("invalid record format"), "[parseBankStatement] error parsing bank statement record")
	}

	amount, err := strconv.ParseFloat(record[1], 64)
	if err != nil {
		return model.BankStatement{}, errors.Wrap(err, "[parseBankStatement] error parsing amount")
	}

	date, err := time.Parse("2006-01-02", record[2])
	if err != nil {
		return model.BankStatement{}, errors.Wrap(err, "[parseBankStatement] error parsing date")
	}

	return model.BankStatement{
		ID:     record[0],
		Amount: amount,
		Date:   date,
	}, nil
}
