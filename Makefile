VERSION=$(shell git rev-parse --verify HEAD)

build:
	VERSION=$(VERSION) docker-compose build

up:
	VERSION=$(VERSION) docker-compose up

down:
	docker-compose down


clean-docker: stop down
	-docker container rm -f demoapi_api demoapi_db demoapi_grafana demoapi_prometheus
	-docker image rm -f samolds/demoapi
	-docker container prune -f
	-docker image prune -f
	-docker volume prune -f
	-docker network prune -f
	-rm -rf monitor/grafana monitor/grafana_data

start:
	docker-compose start

stop:
	docker-compose stop

shell-server: start
	docker exec -ti demoapi_api /bin/sh

shell-db: start
	docker exec -ti demoapi_db /bin/sh
