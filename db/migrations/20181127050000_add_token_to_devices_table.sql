-- migrate:up

ALTER TABLE devices DROP COLUMN public_key;
ALTER TABLE devices ADD COLUMN token VARBINARY(48) NOT NULL UNIQUE KEY;

-- migrate:down

ALTER TABLE devices DROP COLUMN token;
ALTER TABLE devices ADD COLUMN public_key VARBINARY(255) NOT NULL UNIQUE KEY;
