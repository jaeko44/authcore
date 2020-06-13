-- migrate:up
create table audit_logs (
       id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
       actor_id BIGINT,
       action VARCHAR(255) NOT NULL,
       target JSON,
       session_id BIGINT,
       device VARCHAR(255),
       ip VARCHAR(255),
       description VARCHAR(255),
       result INT(1) NOT NULL,
       is_external INT(1) NOT NULL DEFAULT 0,

       created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

       CONSTRAINT FOREIGN KEY (actor_id) REFERENCES users (id),
       CONSTRAINT FOREIGN KEY (session_id) REFERENCES sessions (id)

)

-- migrate:down
drop table audit_logs;
