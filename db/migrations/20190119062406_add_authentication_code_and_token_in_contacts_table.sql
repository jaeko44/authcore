-- migrate:up
ALTER TABLE `contacts` ADD `authentication_info_sent_at` TIMESTAMP after `failed_verification_attempts`;
ALTER TABLE `contacts` ADD `authentication_code` VARCHAR(16) after `authentication_info_sent_at`;
ALTER TABLE `contacts` ADD `authentication_token` VARCHAR(64) after `authentication_code`;
ALTER TABLE `contacts` ADD `failed_authentication_attempts` INT NOT NULL DEFAULT 0 after `authentication_token`;
ALTER TABLE `users` ADD `is_sms_authentication_enabled` INT(1) NOT NULL DEFAULT 0 after `encrypted_password_verifier`;

-- migrate:down
ALTER TABLE `contacts` DROP COLUMN `authentication_info_sent_at`;
ALTER TABLE `contacts` DROP COLUMN `authentication_code`;
ALTER TABLE `contacts` DROP COLUMN `authentication_token`;
ALTER TABLE `contacts` DROP COLUMN `failed_authentication_attempts`;
ALTER TABLE `users` DROP COLUMN `is_sms_authentication_enabled`;
