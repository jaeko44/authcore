-- migrate:up
create table totp_authenticators (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    totp_secret VARCHAR(64),
    status SMALLINT,
    identifier VARCHAR(140),
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT FOREIGN KEY (user_id) REFERENCES users (id)
);

-- migrate:down
drop table totp_authenticators;
