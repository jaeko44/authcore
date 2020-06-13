-- migrate:up
alter table `sessions` add `client_id` VARCHAR(255) after `user_id`;

-- migrate:down
alter table `sessions` drop column `client_id`;
