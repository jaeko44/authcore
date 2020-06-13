-- migrate:up

alter table users
    change column password_salt password_salt VARCHAR(48);

-- migrate:down

alter table users
    change column password_salt password_salt VARCHAR(32);
