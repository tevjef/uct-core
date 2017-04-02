FROM jwilder/docker-gen

COPY nginx.tmpl .

RUN mkdir -p /etc/docker-gen/templates && \
    mv nginx.tmpl /etc/docker-gen/templates/