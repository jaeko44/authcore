-- migrate:up

ALTER TABLE `audit_logs` ADD COLUMN `actor_display` VARCHAR(255) AFTER `actor_id`;
ALTER TABLE `audit_logs` ADD COLUMN `user_agent` VARCHAR(255) AFTER `ip`;

-- migrate:down

ALTER TABLE `audit_logs` DROP COLUMN `actor_display`;
ALTER TABLE `audit_logs` DROP COLUMN `user_agent`;