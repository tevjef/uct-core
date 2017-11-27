-- Disallow new connections
UPDATE pg_database SET datallowconn = 'false' WHERE datname = 'universityct';
ALTER DATABASE universityct CONNECTION LIMIT 1;

-- Terminate existing connections
SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = 'universityct';

-- Drop database
DROP DATABASE universityct;
