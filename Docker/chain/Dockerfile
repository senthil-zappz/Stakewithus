FROM golang:1.19-buster

WORKDIR /chain
COPY . /chain

RUN make install

CMD bandd start --rpc.laddr tcp://0.0.0.0:26657
