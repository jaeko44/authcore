-- migrate:up

alter table sessions
    change column last_seen_ip last_seen_ip VARCHAR(45)

-- migrate:down

alter table sessions
    change column last_seen_ip last_seen_ip VARCHAR(38)
