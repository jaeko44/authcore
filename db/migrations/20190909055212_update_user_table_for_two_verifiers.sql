-- migrate:up
ALTER TABLE `users` DROP column `encrypted_password_verifier`;
ALTER TABLE `users` ADD `encrypted_password_verifier_w0` VARCHAR(1024) after `password_salt`;
ALTER TABLE `users` ADD `encrypted_password_verifier_l` VARCHAR(1024) after `encrypted_password_verifier_w0`;

-- migrate:down
ALTER TABLE `users` DROP column `encrypted_password_verifier_w0`;
ALTER TABLE `users` DROP column `encrypted_password_verifier_l`;
ALTER TABLE `users` ADD `encrypted_password_verifier` VARCHAR(1024) after `password_salt`;
