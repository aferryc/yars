GRANT ALL PRIVILEGES ON DATABASE yars TO postgres;

CREATE TABLE IF NOT EXISTS transactions (
    id VARCHAR(255) PRIMARY KEY,
    amount DECIMAL(15, 2) NOT NULL,
    type VARCHAR(50) NOT NULL,
    transaction_time TIMESTAMP NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS bank_statements (
    id SERIAL PRIMARY KEY,
    amount DECIMAL(15, 2) NOT NULL,
    date TIMESTAMP NOT NULL,
    reference VARCHAR(255),
    bank VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS recon_summary (
    id VARCHAR(255) PRIMARY KEY,
    matched INTEGER NOT NULL,
    discrepancy DECIMAL(15, 2) NOT NULL,
    total_transaction INTEGER NOT NULL,
    total_unmatched_bank INTEGER NOT NULL,
    total_unmatched_internal INTEGER NOT NULL,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS unmatched_transactions (
    id VARCHAR(255) PRIMARY KEY,
    task_id VARCHAR(255) NOT NULL REFERENCES recon_summary(id),
    amount DECIMAL(15, 2) NOT NULL,
    transaction_time TIMESTAMP NOT NULL,
    type VARCHAR(50) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS unmatched_bank_statements (
    id SERIAL PRIMARY KEY,
    task_id VARCHAR(255) NOT NULL REFERENCES recon_summary(id),
    amount DECIMAL(15, 2) NOT NULL,
    date TIMESTAMP NOT NULL,
    reference VARCHAR(255),
    bank_name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_bank_statements_date ON bank_statements(date);
CREATE INDEX IF NOT EXISTS idx_bank_statements_amount ON bank_statements(amount);
CREATE INDEX IF NOT EXISTS idx_bank_statements_bank ON bank_statements(bank);
CREATE UNIQUE INDEX IF NOT EXISTS idx_bank_statements_id_date_bank ON bank_statements(id, date, bank);

CREATE INDEX IF NOT EXISTS idx_transactions_transaction_time ON transactions(transaction_time);
CREATE INDEX IF NOT EXISTS idx_transactions_amount ON transactions(amount);
CREATE INDEX IF NOT EXISTS idx_transactions_type ON transactions(type);
CREATE UNIQUE INDEX IF NOT EXISTS idx_transactions_id_transaction_time ON transactions(id, transaction_time);

CREATE INDEX IF NOT EXISTS idx_recon_summary_date_range ON recon_summary(start_date, end_date);
CREATE INDEX IF NOT EXISTS idx_recon_summary_created_at ON recon_summary(created_at);

CREATE INDEX IF NOT EXISTS idx_unmatched_transactions_task_id ON unmatched_transactions(task_id);
CREATE INDEX IF NOT EXISTS idx_unmatched_transactions_time ON unmatched_transactions(transaction_time);
CREATE INDEX IF NOT EXISTS idx_unmatched_transactions_amount ON unmatched_transactions(amount);
CREATE INDEX IF NOT EXISTS idx_unmatched_transactions_task_time ON unmatched_transactions(task_id, transaction_time);

CREATE INDEX IF NOT EXISTS idx_unmatched_bank_statements_task_id ON unmatched_bank_statements(task_id);
CREATE INDEX IF NOT EXISTS idx_unmatched_bank_statements_date ON unmatched_bank_statements(date);
CREATE INDEX IF NOT EXISTS idx_unmatched_bank_statements_amount ON unmatched_bank_statements(amount);
CREATE INDEX IF NOT EXISTS idx_unmatched_bank_statements_task_date ON unmatched_bank_statements(task_id, date);
CREATE INDEX IF NOT EXISTS idx_unmatched_bank_statements_bank_name ON unmatched_bank_statements(bank_name);

CREATE INDEX IF NOT EXISTS idx_unmatched_txn_task_time_desc ON unmatched_transactions(task_id, transaction_time DESC);
CREATE INDEX IF NOT EXISTS idx_unmatched_bank_task_date_desc ON unmatched_bank_statements(task_id, date DESC);