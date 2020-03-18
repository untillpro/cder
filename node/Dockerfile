FROM node:10.16.1

USER root
RUN mkdir /cder
COPY cder /cder/cder
COPY node/deploy.sh /cder/deploy.sh
RUN chmod 777 /cder/cder
RUN chmod 777 /cder/deploy.sh
RUN apt update
RUN apt install nginx -y
EXPOSE 81
COPY node/nginx/default.conf /etc/nginx/conf.d/default.conf

ENTRYPOINT ["/cder/cder"]