-- +migrate Up

CREATE TABLE `import_history` (
  `id` INT NOT NULL AUTO_INCREMENT comment 'primary id',
  `job_label` TEXT comment 'importer sets joblabel',
  `parsed_entry_num` INT NOT NULL,
  `new_entry_num` INT NOT NULL,
  `created_at` datetime  default current_timestamp,
  `updated_at` timestamp default current_timestamp on update current_timestamp,
  PRIMARY KEY (`id`),
  INDEX `idx1` (`job_label`)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- +migrate Down

DROP TABLE `import_history`;
