-- +migrate Up
CREATE TABLE `sbi_holding` (
    `id` INT NOT NULL AUTO_INCREMENT,
    `snapshot_id` INT NOT NULL COMMENT 'FK -> sbi_snapshot.id (no DB FK, enforced at app layer)',
    `section` VARCHAR(32) NOT NULL COMMENT 'nisa_domestic|nisa_us|nisa_funds|old_nisa_funds',
    `name` TEXT NOT NULL COMMENT 'Holding.name (銘柄名)',
    `quantity` DECIMAL(18,6) NOT NULL COMMENT 'Holding.quantity (口数/株数)',
    `unit_cost` DECIMAL(18,6) NOT NULL COMMENT 'Holding.unit_cost (取得単価, USD for US stocks, JPY for others)',
    `unit_price` DECIMAL(18,6) NOT NULL COMMENT 'Holding.unit_price (現在値)',
    `prev_day_jpy` DECIMAL(14,2) NOT NULL COMMENT 'Holding.prev_day_jpy (前日比円, 0 if unavailable for US)',
    `prev_day_pct` DECIMAL(10,4) NOT NULL COMMENT 'Holding.prev_day_pct (前日比%)',
    `pnl_jpy` DECIMAL(14,2) NOT NULL COMMENT 'Holding.pnl_jpy (評価損益 円)',
    `pnl_pct` DECIMAL(10,4) NOT NULL COMMENT 'Holding.pnl_pct (評価損益%)',
    `value_jpy` DECIMAL(14,2) NOT NULL COMMENT 'Holding.value_jpy (評価額 円)',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    INDEX `idx_snapshot_id` (`snapshot_id`),
    INDEX `idx_snapshot_section` (`snapshot_id`, `section`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +migrate Down
DROP TABLE `sbi_holding`;
