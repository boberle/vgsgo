vgsgo:
	mkdir -p build
	go build -o build/vgsgo cmd/app/main.go

test:
	go test -v ./...
