-- +migrate Up

CREATE TABLE `metrics_fin_ins_update` (
  `fin_ins` TEXT NOT NULL comment 'finance instrcument name',
  `last_date` datetime NOT NULL comment 'record date yyyymm at money-forward',
  `count` NOT NULL INT,
  `maw_count` NOT NULL INT,
  `created_at` datetime  default current_timestamp,
  `updated_at` timestamp current_timestamp on update current_timestamp,
  PRIMARY KEY (`fin_ins`)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- +migrate Down

DROP TABLE `metrics_fin_ins_update`;
