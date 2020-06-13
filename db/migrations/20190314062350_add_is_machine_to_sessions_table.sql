-- migrate:up
ALTER TABLE `sessions` ADD COLUMN `is_machine` tinyint(1) NOT NULL DEFAULT 0;
ALTER TABLE `sessions` CHANGE COLUMN `device_id` `device_id` BIGINT;

-- migrate:down
ALTER TABLE `sessions` DROP COLUMN `is_machine`;
ALTER TABLE `sessions` CHANGE COLUMN `device_id` `device_id` BIGINT NOT NULL;