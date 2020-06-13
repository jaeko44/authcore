-- migrate:up
ALTER TABLE `users` ADD `user_metadata` JSON after `lock_description`;
ALTER TABLE `users` ADD `app_metadata` JSON after `user_metadata`;

-- migrate:down
ALTER TABLE `users` DROP column `user_metadata`;
ALTER TABLE `users` DROP column `app_metadata`;
