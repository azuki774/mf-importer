-- +migrate Up
CREATE TABLE `nrkn_snapshot` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `fetched_at` DATETIME(6) NOT NULL,
    `schema_version` VARCHAR(32) NOT NULL,
    `status` VARCHAR(16) NOT NULL,
    `grand_total_jpy` BIGINT NOT NULL,
    `total_cost_jpy` BIGINT NOT NULL,
    `pnl_jpy` BIGINT NOT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uq_fetched_at` (`fetched_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `nrkn_holding` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `snapshot_id` BIGINT NOT NULL,
    `product_code` VARCHAR(64) COLLATE utf8mb4_bin NOT NULL,
    `composite_figi` VARCHAR(12) COLLATE utf8mb4_bin NOT NULL,
    `name` TEXT NOT NULL,
    `category` TEXT NOT NULL,
    `quantity` DOUBLE NOT NULL,
    `unit_price` DOUBLE NOT NULL,
    `value_jpy` BIGINT NOT NULL,
    `cost_jpy` BIGINT NOT NULL,
    `redemption_unit_price` DOUBLE NOT NULL,
    `redemption_value_jpy` BIGINT NOT NULL,
    `pnl_jpy` BIGINT NOT NULL,
    `reference_date` DATE NOT NULL,
    `allocation_pct` DOUBLE NOT NULL,
    `unit_price_raw` TEXT NOT NULL,
    `redemption_unit_price_raw` TEXT NOT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uq_snapshot_product` (`snapshot_id`, `product_code`),
    KEY `idx_composite_figi` (`composite_figi`),
    CONSTRAINT `fk_nrkn_holding_snapshot` FOREIGN KEY (`snapshot_id`) REFERENCES `nrkn_snapshot` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +migrate Down
DROP TABLE `nrkn_holding`;
DROP TABLE `nrkn_snapshot`;
