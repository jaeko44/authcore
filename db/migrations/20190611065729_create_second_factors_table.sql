-- migrate:up
create table second_factors (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  type SMALLINT NOT NULL,
  content JSON NOT NULL,

  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  last_used_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  CONSTRAINT FOREIGN KEY (user_id) REFERENCES users (id)
);
alter table `users` drop column `is_sms_authentication_enabled`;
drop table totp_authenticators;

-- migrate:down
drop table second_factors;
alter table `users` add column `is_sms_authentication_enabled` tinyint(1) NOT NULL DEFAULT 0;
create table totp_authenticators (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  encrypted_totp_secret VARCHAR(256),
  status SMALLINT,
  identifier VARCHAR(140),
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  last_used_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  CONSTRAINT FOREIGN KEY (user_id) REFERENCES users (id)
);
