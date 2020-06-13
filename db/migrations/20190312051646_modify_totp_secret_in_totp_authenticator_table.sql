-- migrate:up
ALTER TABLE totp_authenticators CHANGE totp_secret encrypted_totp_secret varchar(255);

-- migrate:down

ALTER TABLE totp_authenticators CHANGE encrypted_totp_secret totp_secret varchar(64);