-- +migrate Up

ALTER TABLE `import_history` ADD COLUMN `src_file` TEXT;
ALTER TABLE `import_history` ADD INDEX idx2 (`src_file`);

-- +migrate Down

ALTER TABLE `import_history` DROP COLUMN `src_file`;
