-- migrate:up
create table roles (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(140) NOT NULL UNIQUE,
    is_system_role INT(1) NOT NULL DEFAULT 0,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

create table roles_users (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    role_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,

    CONSTRAINT FOREIGN KEY (role_id) REFERENCES roles (id),
    CONSTRAINT FOREIGN KEY (user_id) REFERENCES users (id)
);

insert into roles (name, is_system_role) values ("authcore.admin", 1);
insert into roles (name, is_system_role) values ("authcore.editor", 1);

-- migrate:down
drop table roles_users;
drop table roles;
