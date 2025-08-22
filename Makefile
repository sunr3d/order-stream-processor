env:
	if [ ! -f .env ]; then cp .env.example .env; fi

up: env
	docker compose up -d --build

down:
	docker compose down

restart: down up

clean:
	docker compose down -v
	rm -f .env

logs:
	docker compose logs -f app

test:
	go test -v ./...

test-kafka:
	./test-kafka.sh

demo:
	./demo.sh

fmt:
	go fmt ./...

build:
	go build -o order-stream-processor cmd/main.go