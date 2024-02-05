FROM golang:1.19
WORKDIR /chain
#RUN git clone https://github.com/bandprotocol/chain
#RUN cd chain 
#RUN git checkout v2.5.4
#RUN mkdir chain
COPY ./chain /chain
RUN make install 
#RUN ls -lh
#RUN cd ./cmd/bandd
#RUN go build -o /chain/bin/
#RUN cd /chain/cmd/yoda
#RUN go build -o ../../bin/
#RUN cd /chain/cmd/bandevmbot
#RUN go build -o ../../bin/

CMD ["bandd"]

