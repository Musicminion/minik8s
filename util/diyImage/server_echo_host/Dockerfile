FROM ubuntu:20.04

WORKDIR /app

EXPOSE 8090

COPY ./server_echo_host ./server_echo_host

RUN chmod +x ./server_echo_host


ENTRYPOINT ["/app/server_echo_host"]

