
build:
	go build -o bin/corners main.go

clean:
	rm -rf bin

run: build
	bin/corners