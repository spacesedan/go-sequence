build:
	@go build  -o bin/server ./cmd/rest/main.go
	@./bin/server

air:
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
	make -j5 air watch-bundle watch-tw

docker-up:
	docker compose up -d

docker-down:
	docker compose down -v

