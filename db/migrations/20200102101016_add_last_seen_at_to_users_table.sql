-- migrate:up
ALTER TABLE `users` ADD `last_seen_at` TIMESTAMP NOT NULL DEFAULT '1970-01-01 00:00:01' after `reset_password_count`;

-- migrate:down
ALTER TABLE `users` DROP column `last_seen_at`;
