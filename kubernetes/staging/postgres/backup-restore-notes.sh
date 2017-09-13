
pg_dump --host=A -d dbname -v -w -c -Z 9 > dbname.sql.gz

dropdb --host=B dbname

createdb --host=B dbname

cat dbname.sql.gz | gunzip |  psql --host=B -d dbname
