FROM ubuntu:20.04

WORKDIR /app

EXPOSE 8090

COPY ./autoScale ./autoScale

RUN chmod +x ./autoScale


ENTRYPOINT ["/app/autoScale"]

