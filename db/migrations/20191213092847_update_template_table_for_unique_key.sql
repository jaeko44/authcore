-- migrate:up
ALTER TABLE `templates` ADD UNIQUE KEY `template_key` (`name`, `language`);

-- migrate:down
ALTER TABLE `templates` DROP INDEX `template_key`;