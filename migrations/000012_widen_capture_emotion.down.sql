-- Values longer than 10 characters will fail this down migration
ALTER TABLE captures ALTER COLUMN emotion TYPE VARCHAR(10);
