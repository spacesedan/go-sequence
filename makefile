build:
	@go build  -o bin/server ./cmd/rest/main.go
	@./bin/server

watch-tw:
	npx tailwindcss -i ./assets/css/input.css -o ./assets/css/output.css --watch >> /dev/null
