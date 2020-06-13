-- migrate:up
ALTER TABLE `users` ADD `email` VARCHAR(256) after `encrypted_password_verifier_l`;
ALTER TABLE `users` ADD `email_verified_at` timestamp after `email`;
ALTER TABLE `users` ADD `phone` VARCHAR(256) after `email_verified_at`;
ALTER TABLE `users` ADD `phone_verified_at` timestamp after `phone`;
ALTER TABLE `users` ADD `recovery_email` VARCHAR(256) after `phone_verified_at`;
ALTER TABLE `users` ADD `recovery_email_verified_at` timestamp after `recovery_email`;

-- migrate:down
ALTER TABLE `users` DROP COLUMN `email`;
ALTER TABLE `users` DROP COLUMN `email_verified_at`;
ALTER TABLE `users` DROP COLUMN `phone`;
ALTER TABLE `users` DROP COLUMN `phone_verified_at`;
ALTER TABLE `users` DROP COLUMN `recovery_email`;
ALTER TABLE `users` DROP COLUMN `recovery_email_verified_at`;