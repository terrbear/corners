
build: build_client build_server build_windows

clean:
	rm -rf bin

build_client:
	go build -o bin/corners client/main.go

build_server:
	go build -o bin/server server/main.go

publish:
	scp bin/corners hulk.local:/usr/local/bin

run: build
	bin/server &
	bin/corners &

build_windows:
	GOOS=windows go build -o bin/windows/corners.exe client/main.go