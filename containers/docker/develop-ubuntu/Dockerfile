FROM ubuntu:xenial-20210416

ENV PATH=/usr/lib/go-1.9/bin:$PATH

RUN \
  apt-get update && apt-get upgrade -q -y && \
  apt-get install -y --no-install-recommends golang-1.9 git make gcc libc-dev ca-certificates && \
  git clone --depth 1 https://github.com/AbeyFoundation/go-abey && \
  (cd abey && make gabey) && \
  cp abey/build/bin/gabey /gabey && \
  apt-get remove -y golang-1.9 git make gcc libc-dev && apt autoremove -y && apt-get clean && \
  rm -rf /abey

EXPOSE 8545
EXPOSE 30303

ENTRYPOINT ["/gabey"]
