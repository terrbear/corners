GO ?=go
GO_SERVER ?=${GO}
GO_CLIENT ?=${GO}

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
	rsync -avr maps ubuntu@corners.terrbear.io:
	scp bin/linux/server ubuntu@corners.terrbear.io:

run_dev_server:
	LOG_LEVEL=debug LOBBY_TIMEOUT=0 ${GO_SERVER} run server/main.go

run_dev_client:
	LOG_LEVEL=debug GAME_HOST=localhost:8080 ${GO_CLIENT} run client/main.go

run: build
	bin/server &
	bin/corners &
