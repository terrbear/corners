
build: build_client build_server 

clean:
	rm -rf bin

build_client:
	go build -o bin/corners client/main.go
	GOOS=windows go build -o bin/windows/corners.exe client/main.go

build_server:
	go build -o bin/server server/main.go
	GOOS=linux CGO_ENABLED=0 go build -o bin/linux/server server/main.go

lint:
	golangci-lint run --tests=false ./...

publish:
	scp bin/linux/server ubuntu@corners.terrbear.io:
	rsync -avr maps ubuntu@corners.terrbear.io:

run_local_server:
	LOG_LEVEL=debug LOBBY_TIMEOUT=0 go run server/main.go

run: build
	bin/server &
	bin/corners &