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

watch:
	make -j4 generate watch-tw watch-bundle
