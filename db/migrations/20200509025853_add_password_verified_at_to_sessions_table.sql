-- migrate:up

ALTER TABLE `sessions` ADD COLUMN `last_password_verified_at` timestamp

-- migrate:down

ALTER TABLE `sessions` DROP COLUMN `last_password_verified_at`;
