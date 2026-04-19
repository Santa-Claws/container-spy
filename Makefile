.PHONY: build build-image run-tui run-web clean

build:
	go build -o container-spy ./cmd/container-spy

build-image:
	docker build -t container-spy:latest .

run-tui:
	docker compose run --rm container-spy

run-web:
	CONTAINER_SPY_MODE=web docker compose up -d

clean:
	docker compose down
	rm -f container-spy
