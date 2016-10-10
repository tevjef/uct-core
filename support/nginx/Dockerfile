FROM nginx:alpine

EXPOSE 8090

COPY nginx.conf /etc/nginx/nginx.conf

COPY *.conf /etc/nginx/conf.d/

COPY htpasswd/ /etc/nginx/htpasswd/