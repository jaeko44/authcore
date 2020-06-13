-- migrate:up

create table secrets (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    type SMALLINT NOT NULL,
    encrypted_secret VARCHAR(256) NOT NULL,
    is_exported INT(1) NOT NULL DEFAULT 0,

    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT FOREIGN KEY (user_id) REFERENCES users (id)
);

-- migrate:down
drop table secrets;
