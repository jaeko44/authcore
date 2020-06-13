-- migrate:up
ALTER TABLE `users` ADD `is_locked` INT(1) NOT NULL DEFAULT 0 after `is_sms_authentication_enabled`;
ALTER TABLE `users` ADD `lock_expired_at` TIMESTAMP after `is_locked`;
ALTER TABLE `users` ADD `lock_description` VARCHAR(256) after `lock_expired_at`;

-- migrate:down
ALTER TABLE `users` DROP column `is_locked`;
ALTER TABLE `users` DROP column `lock_expired_at`;
ALTER TABLE `users` DROP column `lock_description`;