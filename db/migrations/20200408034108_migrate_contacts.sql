-- migrate:up

UPDATE users,
       (SELECT user_id, `value`, verified_at
        FROM contacts
        WHERE type=0 AND is_primary=1)
        AS primary_email
SET users.email = primary_email.value, users.email_verified_at = primary_email.verified_at
WHERE primary_email.user_id = users.id;

UPDATE users,
       (SELECT user_id, `value`, verified_at
        FROM contacts
        WHERE type=1 AND is_primary=1)
        AS primary_phone
SET users.phone = primary_phone.value, users.phone_verified_at = primary_phone.verified_at
WHERE primary_phone.user_id = users.id;

CREATE UNIQUE INDEX email ON users (email);
CREATE UNIQUE INDEX phone ON users (phone);
DROP INDEX username ON users;

-- migrate:down

DROP INDEX email ON users;
DROP INDEX phone ON users;
CREATE UNIQUE INDEX username ON users (username);