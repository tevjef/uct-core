# Dump gzipped database
PGPASSWORD=passs pg_dump --host=dbtobackup -U universityct -v -w -c -Z 9 > universityct.sql.gz
# Disallow new connections on the database
echo -n "UPDATE pg_database SET datallowconn = 'false' WHERE datname = 'universityct';" |  psql --host=localhost -U universityct -d postgres
# Drop all exisiting connections on the database
echo -n "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = 'universityct';" |  psql --host=localhost -U universityct -d postgres
# Restore the database
cat universityct.sql.gz | gunzip |  psql --host=localhost -U universityct -d universityct
# Allow connections to the database
echo -n "UPDATE pg_database SET datallowconn = 'true' WHERE datname = 'universityct';" |  psql --host=localhost -U universityct -d postgres