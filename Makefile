export CGO_ENABLED = 0

build: server zxing

server:
	go build -v -mod=mod \
	-o bin/server \
	-buildvcs=false \
	mavtrainticketscandemo/cmd/server


zxing:
	cd zxing && make
