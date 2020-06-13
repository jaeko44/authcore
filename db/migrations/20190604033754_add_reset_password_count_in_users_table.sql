-- migrate:up
ALTER TABLE `users` ADD `reset_password_count` BIGINT NOT NULL DEFAULT 0 after `app_metadata`;

-- migrate:down
ALTER TABLE `users` DROP column `reset_password_count`;
