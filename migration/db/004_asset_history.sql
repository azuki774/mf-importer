-- +migrate Up
CREATE TABLE `asset_history` (
    `id` INT NOT NULL AUTO_INCREMENT,
    `date` DATE NOT NULL,
    `total_amount` INT NOT NULL,
    `cash_deposit` INT NOT NULL COMMENT '現金・預金・投資信託',
    `bonds` INT NOT NULL COMMENT '債券',
    `other_assets` INT NOT NULL COMMENT 'その他金融資産',
    `points` INT NOT NULL COMMENT 'ポイント',
    `details` TEXT COMMENT '詳細',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `unique_date` (`date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- +migrate Down
DROP TABLE `asset_history`;