-- migrate:up
ALTER TABLE `oauth_factors` ADD COLUMN `metadata` JSON after `oauth_user_id`;

-- migrate:down
ALTER TABLE `oauth_factors` DROP COLUMN `metadata`;