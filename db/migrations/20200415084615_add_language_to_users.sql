-- migrate:up

ALTER TABLE `users` ADD COLUMN `language` VARCHAR(255) AFTER `reset_password_count`;

-- migrate:down

ALTER TABLE `users` DROP COLUMN `language`;
