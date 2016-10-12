FROM nginx:alpine

RUN apk add --update bash && rm -rf /var/cache/apk/*

EXPOSE 8090

COPY nginx.conf /etc/nginx/nginx.conf

COPY htpasswd/ /etc/nginx/htpasswd/