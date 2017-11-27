PGPASSWORD=passs pg_dump --host=foreman.tevindev.me -U universityct -v -w -c -Z 9 > universityct.sql.gz

cat universityct.sql.gz | gunzip |  psql --host=localhost -U universityct -d universityct