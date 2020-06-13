-- migrate:up
alter table users add email VARCHAR(255) after username;

-- migrate:down
alter table users drop column email;