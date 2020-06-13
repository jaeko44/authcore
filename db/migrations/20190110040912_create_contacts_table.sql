-- migrate:up
create table contacts (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    type SMALLINT NOT NULL,
    value VARCHAR(256) NOT NULL,
    is_primary INT(1) NOT NULL DEFAULT 0,

    verification_info_sent_at TIMESTAMP,
    verification_code VARCHAR(16),
    verification_token VARCHAR(64),
    failed_verification_attempts INT NOT NULL DEFAULT 0,

    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    verified_at TIMESTAMP,

    CONSTRAINT FOREIGN KEY (user_id) REFERENCES users (id)
);
alter table users drop column email;

-- migrate:down
drop table contacts;
alter table users add email VARCHAR(255) after username;