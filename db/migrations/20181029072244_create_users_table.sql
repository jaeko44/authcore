-- migrate:up
create table users (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) UNIQUE KEY,
    email VARCHAR(255) UNIQUE KEY,
    phone VARCHAR(50) UNIQUE KEY,
    display_name VARCHAR(50) NOT NULL DEFAULT '',
    password_salt VARCHAR(32),
    encrypted_password_verifier VARCHAR(1024),
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- migrate:down
drop table users;
