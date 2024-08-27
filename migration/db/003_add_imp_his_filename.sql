-- +migrate Up

ALTER TABLE `import_history` ADD COLUMN `src_file` TEXT;

-- +migrate Down

ALTER TABLE `import_history` DROP COLUMN `src_file`;
