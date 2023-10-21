build:
	mkdir -p build
	go build -o build/vgsplay cmd/app/main.go

test:
	go test -v ./...
