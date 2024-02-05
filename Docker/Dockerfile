FROM golang:1.19
WORKDIR /chain

COPY ./chain /chain
RUN make install 

CMD ["bandd"]

