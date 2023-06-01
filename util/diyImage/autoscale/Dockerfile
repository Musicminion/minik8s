FROM ubuntu:20.04

WORKDIR /app

EXPOSE 8080

COPY ./autoscale ./autoscale

RUN chmod +x ./autoscale


ENTRYPOINT ["/app/autoscale"]

