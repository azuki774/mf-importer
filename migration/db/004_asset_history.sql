-- +migrate Up
CREATE TABLE `asset_history` (
    `id` INT NOT NULL AUTO_INCREMENT,
    `date` DATE NOT NULL,
    `total_amount` INT NOT NULL COMMENT '合計',
    `cash_deposit_crypto` INT NOT NULL COMMENT '預金・現金・暗号資産',
    `stocks` INT NOT NULL COMMENT '株式(現物)',
    `investment_trusts` INT NOT NULL COMMENT '投資信託',
    `points` INT NOT NULL COMMENT 'ポイント',
    `details` TEXT COMMENT '詳細',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `unique_date` (`date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- +migrate Down
DROP TABLE `asset_history`;
