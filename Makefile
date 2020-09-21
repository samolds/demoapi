VERSION=$(shell git rev-parse --verify HEAD)

build:
	VERSION=$(VERSION) docker-compose build

up:
	docker-compose up

clean-docker: stop
	-docker container rm -f demoapi_api demoapi_db demoapi_grafana demoapi_prometheus
	-docker image rm -f demoapi_api demoapi_db demoapi_grafana demoapi_prometheus
	-docker container prune -f
	-docker image prune -f

start:
	docker-compose start

stop:
	docker-compose stop

shell-server: start
	docker exec -ti demoapi_api /bin/sh

shell-db: start
	docker exec -ti demoapi_db /bin/sh
