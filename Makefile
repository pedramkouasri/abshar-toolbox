build:
	CGO_ENABLED=0 GOOS=linux go build -o bin/abshar-toolbox main.go

build-server: build
	scp bin/abshar-toolbox root@10.10.10.226:/home/pedram/bin

create:
	go run main.go patch create