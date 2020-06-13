-- migrate:up

ALTER TABLE `sessions` DROP COLUMN `credential_type`;
ALTER TABLE `sessions` DROP COLUMN `credential`;
ALTER TABLE `sessions` ADD COLUMN `refresh_token` varbinary(48) NOT NULL UNIQUE KEY;

-- migrate:down

ALTER TABLE `sessions` DROP COLUMN `refresh_token`;
ALTER TABLE `sessions` ADD COLUMN `credential` varchar(255) NOT NULL;
ALTER TABLE `sessions` ADD COLUMN `credential_type` smallint(6) NOT NULL;