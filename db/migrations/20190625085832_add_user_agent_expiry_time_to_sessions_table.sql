-- migrate:up

ALTER TABLE `sessions` ADD COLUMN `user_agent` VARCHAR(255) default "";
ALTER TABLE `sessions` ADD COLUMN `expired_at` TIMESTAMP;
ALTER TABLE `sessions` ADD COLUMN `is_invalid` TINYINT(1) NOT NULL DEFAULT 0;

-- migrate:down

ALTER TABLE `sessions` DROP COLUMN `user_agent`;
ALTER TABLE `sessions` DROP COLUMN `expired_at`;
ALTER TABLE `sessions` DROP COLUMN `is_invalid`;
