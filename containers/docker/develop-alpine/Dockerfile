FROM alpine:3.14

RUN \
  apk add --update go git make gcc musl-dev=1.1.19-r10 linux-headers ca-certificates && \
  git clone --depth 1 https://github.com/AbeyFoundation/go-abey && \
  (cd abey && make gabey) && \
  cp abey/build/bin/gabey /gabey && \
  apk del go git make gcc musl-dev linux-headers && \
  rm -rf /abey && rm -rf /var/cache/apk/*

EXPOSE 8545
EXPOSE 30303

ENTRYPOINT ["/gabey"]
