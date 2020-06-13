-- migrate:up
ALTER TABLE audit_logs DROP FOREIGN KEY audit_logs_ibfk_1;
ALTER TABLE audit_logs DROP FOREIGN KEY audit_logs_ibfk_2;

-- migrate:down
ALTER TABLE audit_logs ADD CONSTRAINT audit_logs_ibfk_1 FOREIGN KEY (`actor_id`) REFERENCES `users` (`id`);
ALTER TABLE audit_logs ADD CONSTRAINT audit_logs_ibfk_2 FOREIGN KEY (`session_id`) REFERENCES `sessions` (`id`);
