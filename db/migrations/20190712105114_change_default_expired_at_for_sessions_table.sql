-- migrate:up
UPDATE `sessions` SET expired_at = NOW();
ALTER TABLE `sessions` MODIFY expired_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

-- migrate:down

ALTER TABLE `sessions` MODIFY expired_at TIMESTAMP DEFAULT NULL;