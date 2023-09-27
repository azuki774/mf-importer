-- +migrate Up

CREATE TABLE `detail` (
  `id` INT NOT NULL AUTO_INCREMENT comment 'primary id',
  `yyyymm_id` INT NOT NULL comment 'id for each yyyymm',
  `date` DATE NOT NULL comment 'record date yyyymm',
  `name` TEXT comment 'detail name',
  `price` INT,
  `fin_ins` TEXT comment 'finance instrcument name',
  `l_category` TEXT comment 'large category name',
  `m_category` TEXT comment 'medium category name',
  `regist_date` DATE NOT NULL comment 'date running importer',
  `maw_check_date` DATE comment 'mawinter check date',
  `maw_regist_date` DATE comment 'mawinter regist check date',
  `raw_date` TEXT,
  `raw_price` TEXT,
  `created_at` datetime  default current_timestamp,
  `updated_at` timestamp default current_timestamp on update current_timestamp,
  PRIMARY KEY (`id`),
  INDEX `idx1` (`maw_check_date`),
  INDEX `idx2` (`name`),
  INDEX `idx3` (`raw_price`)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

CREATE TABLE `extract_rule` (
  `id` INT NOT NULL AUTO_INCREMENT comment 'primary id',
  `field_name` TEXT NOT NULL comment 'extract field name (m_category or name)',
  `value` TEXT NOT NULL,
  `exact_match` INT comment 'exact match = 1 or not 0',
  `category_id` INT NOT NULL comment 'mawinter category id',
  `created_at` datetime  default current_timestamp,
  `updated_at` timestamp default current_timestamp on update current_timestamp,
  PRIMARY KEY (`id`)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- +migrate Down

DROP TABLE `detail`;
DROP TABLE `extract_rule`;
