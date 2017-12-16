FROM postgres:9.6

COPY . /docker-entrypoint-initdb.d/

EXPOSE 5432

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["postgres"]
