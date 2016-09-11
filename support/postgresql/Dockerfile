FROM postgres:9.5.4

COPY . /docker-entrypoint-initdb.d/

ENTRYPOINT ["/docker-entrypoint.sh"]

ENV POSTGRES_USER=uct
ENV POSTGRES_PASSWORD=***REMOVED***
ENV POSTGRES_DB=universityct

EXPOSE 5432
CMD ["postgres"]
