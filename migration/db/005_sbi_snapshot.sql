-- +migrate Up
CREATE TABLE `sbi_snapshot` (
    `id` INT NOT NULL AUTO_INCREMENT,
    `fetched_at` DATETIME(6) NOT NULL COMMENT 'Assets.fetched_at (UTC, nanoseconds preserved)',
    `status` VARCHAR(16) NOT NULL COMMENT 'ok|maintenance',
    `schema_version` INT NOT NULL COMMENT 'Assets.schema_version (CurrentSchemaVersion)',
    `grand_total_jpy` DECIMAL(14,2) NOT NULL COMMENT 'grand_total_jpy = nisa+old_nisa+cash+others',
    -- NISA summary (new NISA)
    `nisa_total_jpy` DECIMAL(14,2) NOT NULL COMMENT 'nisa.total_jpy',
    `nisa_prev_day_jpy` DECIMAL(14,2) NOT NULL COMMENT 'nisa.prev_day_jpy',
    `nisa_prev_day_pct` DECIMAL(10,4) NOT NULL COMMENT 'nisa.prev_day_pct',
    `nisa_prev_month_jpy` DECIMAL(14,2) NOT NULL COMMENT 'nisa.prev_month_jpy',
    `nisa_prev_month_pct` DECIMAL(10,4) NOT NULL COMMENT 'nisa.prev_month_pct',
    `nisa_pnl_jpy` DECIMAL(14,2) NOT NULL COMMENT 'nisa.pnl_jpy (評価損益)',
    `nisa_pnl_pct` DECIMAL(10,4) NOT NULL COMMENT 'nisa.pnl_pct',
    -- NISA domestic_stocks
    `nisa_domestic_value_jpy` DECIMAL(14,2) NOT NULL COMMENT 'nisa.domestic_stocks.value_jpy',
    `nisa_domestic_pnl_jpy` DECIMAL(14,2) NOT NULL COMMENT 'nisa.domestic_stocks.pnl_jpy',
    `nisa_domestic_pnl_pct` DECIMAL(10,4) NOT NULL COMMENT 'nisa.domestic_stocks.pnl_pct',
    `nisa_domestic_prev_day_jpy` DECIMAL(14,2) NOT NULL COMMENT 'nisa.domestic_stocks.prev_day_jpy',
    `nisa_domestic_prev_day_pct` DECIMAL(10,4) NOT NULL COMMENT 'nisa.domestic_stocks.prev_day_pct',
    `nisa_domestic_prev_month_jpy` DECIMAL(14,2) NOT NULL COMMENT 'nisa.domestic_stocks.prev_month_jpy',
    `nisa_domestic_prev_month_pct` DECIMAL(10,4) NOT NULL COMMENT 'nisa.domestic_stocks.prev_month_pct',
    -- NISA us_stocks
    `nisa_us_value_jpy` DECIMAL(14,2) NOT NULL COMMENT 'nisa.us_stocks.value_jpy',
    `nisa_us_pnl_jpy` DECIMAL(14,2) NOT NULL COMMENT 'nisa.us_stocks.pnl_jpy',
    `nisa_us_pnl_pct` DECIMAL(10,4) NOT NULL COMMENT 'nisa.us_stocks.pnl_pct',
    `nisa_us_prev_day_jpy` DECIMAL(14,2) NOT NULL COMMENT 'nisa.us_stocks.prev_day_jpy (0 for US, no prev-day on foreign page)',
    `nisa_us_prev_day_pct` DECIMAL(10,4) NOT NULL COMMENT 'nisa.us_stocks.prev_day_pct',
    `nisa_us_prev_month_jpy` DECIMAL(14,2) NOT NULL COMMENT 'nisa.us_stocks.prev_month_jpy',
    `nisa_us_prev_month_pct` DECIMAL(10,4) NOT NULL COMMENT 'nisa.us_stocks.prev_month_pct',
    -- NISA funds
    `nisa_funds_value_jpy` DECIMAL(14,2) NOT NULL COMMENT 'nisa.funds.value_jpy',
    `nisa_funds_pnl_jpy` DECIMAL(14,2) NOT NULL COMMENT 'nisa.funds.pnl_jpy',
    `nisa_funds_pnl_pct` DECIMAL(10,4) NOT NULL COMMENT 'nisa.funds.pnl_pct',
    `nisa_funds_prev_day_jpy` DECIMAL(14,2) NOT NULL COMMENT 'nisa.funds.prev_day_jpy',
    `nisa_funds_prev_day_pct` DECIMAL(10,4) NOT NULL COMMENT 'nisa.funds.prev_day_pct',
    `nisa_funds_prev_month_jpy` DECIMAL(14,2) NOT NULL COMMENT 'nisa.funds.prev_month_jpy',
    `nisa_funds_prev_month_pct` DECIMAL(10,4) NOT NULL COMMENT 'nisa.funds.prev_month_pct',
    -- Old NISA (旧つみたてNISA)
    `old_nisa_total_jpy` DECIMAL(14,2) NOT NULL COMMENT 'old_nisa.total_jpy',
    `old_nisa_prev_day_jpy` DECIMAL(14,2) NOT NULL COMMENT 'old_nisa.prev_day_jpy',
    `old_nisa_prev_day_pct` DECIMAL(10,4) NOT NULL COMMENT 'old_nisa.prev_day_pct',
    `old_nisa_pnl_jpy` DECIMAL(14,2) NOT NULL COMMENT 'old_nisa.pnl_jpy',
    `old_nisa_pnl_pct` DECIMAL(10,4) NOT NULL COMMENT 'old_nisa.pnl_pct',
    -- Cash (cash.jpy / cash.usd)
    `cash_jpy_amount` DECIMAL(14,2) NOT NULL COMMENT 'cash.jpy.amount (== value_jpy for JPY)',
    `cash_jpy_value_jpy` DECIMAL(14,2) NOT NULL COMMENT 'cash.jpy.value_jpy',
    `cash_usd_amount` DECIMAL(14,4) NOT NULL COMMENT 'cash.usd.amount (USD)',
    `cash_usd_value_jpy` DECIMAL(14,2) NOT NULL COMMENT 'cash.usd.value_jpy (JPY converted)',
    -- Others (others.funds = 特定預り 投資信託)
    `other_funds_amount` DECIMAL(14,2) NOT NULL COMMENT 'others.funds.amount',
    `other_funds_value_jpy` DECIMAL(14,2) NOT NULL COMMENT 'others.funds.value_jpy',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uq_fetched_at` (`fetched_at`),
    INDEX `idx_status` (`status`),
    INDEX `idx_fetched_at_desc` (`fetched_at` DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +migrate Down
DROP TABLE `sbi_snapshot`;
