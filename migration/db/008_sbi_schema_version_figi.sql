-- +migrate Up
ALTER TABLE `sbi_snapshot`
    MODIFY `schema_version` VARCHAR(32) NOT NULL COMMENT 'Assets.schema_version (date-based schema version)';

ALTER TABLE `sbi_holding`
    ADD COLUMN `composite_figi` VARCHAR(12) NULL COMMENT 'Holding.composite_figi (nullable for holdings imported before this migration)',
    ADD UNIQUE KEY `uq_snapshot_section_figi` (`snapshot_id`, `section`, `composite_figi`),
    ADD INDEX `idx_composite_figi` (`composite_figi`);

-- +migrate Down
-- The checks deliberately fail with a duplicate-key error when rollback would
-- lose data. This is conditional and works even when strict SQL mode is off.
CREATE TEMPORARY TABLE `_mf_importer_008_down_guard` (
    `guard_id` TINYINT NOT NULL PRIMARY KEY
);

INSERT INTO `_mf_importer_008_down_guard` (`guard_id`)
SELECT 1
FROM `sbi_snapshot`
WHERE `schema_version` NOT REGEXP '^-?[0-9]+$'
   OR CAST(`schema_version` AS DECIMAL(38,0)) < -2147483648
   OR CAST(`schema_version` AS DECIMAL(38,0)) > 2147483647
UNION ALL
SELECT 1
FROM `sbi_snapshot`
WHERE `schema_version` NOT REGEXP '^-?[0-9]+$'
   OR CAST(`schema_version` AS DECIMAL(38,0)) < -2147483648
   OR CAST(`schema_version` AS DECIMAL(38,0)) > 2147483647
UNION ALL
SELECT 2
FROM `sbi_holding`
WHERE `composite_figi` IS NOT NULL
UNION ALL
SELECT 2
FROM `sbi_holding`
WHERE `composite_figi` IS NOT NULL;

DROP TEMPORARY TABLE `_mf_importer_008_down_guard`;

ALTER TABLE `sbi_holding`
    DROP INDEX `uq_snapshot_section_figi`,
    DROP INDEX `idx_composite_figi`,
    DROP COLUMN `composite_figi`;

ALTER TABLE `sbi_snapshot`
    MODIFY `schema_version` INT NOT NULL COMMENT 'Assets.schema_version (CurrentSchemaVersion)';
