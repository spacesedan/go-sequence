build:
	@go build  -o bin/server ./cmd/rest/main.go
	@./bin/server

generate:
	air

watch-tw:
	npx tailwindcss -i ./src/input.css -o ./src/output.css --watch

watch-bundle:
	npm run build:watch

dev:
	npm run dev

bundle:
	npm run build

watch: watch-tw watch-bundle
