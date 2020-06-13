-- migrate:up
ALTER TABLE `contacts` DROP COLUMN `verification_code`;
ALTER TABLE `contacts` DROP COLUMN `verification_token`;
ALTER TABLE `contacts` DROP COLUMN `verification_info_sent_at`;
ALTER TABLE `contacts` DROP COLUMN `failed_verification_attempts`;
ALTER TABLE `contacts` DROP COLUMN `authentication_code`;
ALTER TABLE `contacts` DROP COLUMN `authentication_token`;
ALTER TABLE `contacts` DROP COLUMN `authentication_info_sent_at`;
ALTER TABLE `contacts` DROP COLUMN `failed_authentication_attempts`;

-- migrate:down
ALTER TABLE `contacts` ADD `verification_code` VARCHAR(16) after `is_primary`;
ALTER TABLE `contacts` ADD `verification_token` VARCHAR(64) after `verification_code`;
ALTER TABLE `contacts` ADD `verification_info_sent_at` VARCHAR(16) after `verification_token`;
ALTER TABLE `contacts` ADD `failed_verification_attempts` INT NOT NULL DEFAULT 0 after `verification_token`;
ALTER TABLE `contacts` ADD `authentication_code` VARCHAR(16) after `failed_verification_attempts`;
ALTER TABLE `contacts` ADD `authentication_token` VARCHAR(64) after `authentication_code`;
ALTER TABLE `contacts` ADD `authentication_info_sent_at` VARCHAR(16) after `authentication_token`;
ALTER TABLE `contacts` ADD `failed_authentication_attempts` INT NOT NULL DEFAULT 0 after `authentication_info_sent_at`;
