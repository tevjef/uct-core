FROM jwilder/docker-gen

RUN wget -q https://raw.githubusercontent.com/jwilder/nginx-proxy/master/nginx.tmpl && \
    sed -i -e 's/vhost;/json;/g' nginx.tmpl && \
    sed -i '/gzip_types/d' nginx.tmpl && \
    mkdir -p /etc/docker-gen/templates && \
    mv nginx.tmpl /etc/docker-gen/templates/