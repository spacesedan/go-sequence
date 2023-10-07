build:
	@go build  -o bin/server ./cmd/rest/main.go
	@./bin/server
