FROM golang:1.17.0

RUN mkdir /cder
COPY cder /cder/cder
RUN chmod 777 /cder/cder

ENTRYPOINT ["/cder/cder"]
