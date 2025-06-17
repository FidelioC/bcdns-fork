FROM ubuntu:24.04

ENV RUST_BACKTRACE 1

WORKDIR /substrate
COPY ./target/release/node-template /substrate/target/release/
COPY ./entrypoint.sh /substrate/
COPY ./customSpecRaw.json /substrate/

RUN apt-get update
RUN apt-get install -y autoconf automake libtool curl make g++ unzip clang
RUN apt-get install -y protobuf-compiler

RUN useradd -m -u 1001 -U -s /bin/sh -d /substrate substrate && \
	mkdir -p /data /substrate/.local/share/substrate && \
	chown -R substrate:substrate /data && \
	ln -s /data /substrate/.local/share/substrate && \
	chmod +x ./entrypoint.sh

USER substrate
EXPOSE 30333 9933 9944 9945 9615
VOLUME ["/data"]

ENTRYPOINT ["./entrypoint.sh"]