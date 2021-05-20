FROM golang:1.16.4

RUN mkdir /cder
COPY cder /cder/cder
RUN chmod 777 /cder/cder

ENTRYPOINT ["/cder/cder"]
