-- migrate:up
alter table users drop column phone;
alter table users drop column email;

-- migrate:down
alter table users add email VARCHAR(255) after username;
alter table users add phone VARCHAR(50) after email;
