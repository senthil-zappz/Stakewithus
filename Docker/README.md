# Method 02

Create a docker image that runs the following project binary using Dockefile. It must be able to build locally.

https://github.com/bandprotocol/chain

Version to build: v2.5.4
```
brew install go
git clone https://github.com/bandprotocol/chain
cd chain
git checkout v2.5.4
make install
cd cmd/bandd
go build -o ../../bin/
cd ../yoda
go build -o ../../bin/
cd ../bandevmbot
go build -o ../../bin/
```