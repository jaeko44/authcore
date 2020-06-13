-- migrate:up

ALTER TABLE `users` ADD COLUMN `name` VARCHAR(255) AFTER `username`;

-- migrate:down

ALTER TABLE `users` DROP COLUMN `name`;