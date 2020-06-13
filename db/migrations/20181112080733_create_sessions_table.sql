-- migrate:up
create table `sessions` (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    device_id BIGINT NOT NULL,
    credential_type SMALLINT NOT NULL,
    `credential` VARCHAR(255) NOT NULL,

    last_seen_at TIMESTAMP,
    last_seen_location VARCHAR(255),
    last_seen_ip VARCHAR(38),

    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE KEY(credential_type, `credential`),
    CONSTRAINT FOREIGN KEY (user_id) REFERENCES users (id)
);

-- migrate:down
drop table `sessions`;