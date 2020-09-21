VERSION ?= alpha

# TODO(sam): dbx is no longer stable. go get .../dbx.v1 fails due to unfound
# dependencies, so the docker image wont be able to use "make setup" and have a
# dbx binary to generate fresh. so the generate calls can only run during dev.

all: server

server-no-generate:
	go build -ldflags "-X main.version=$(VERSION)" -o ./demoapi .

server: generate
	make server-no-generate

setup-no-dbx:
	go mod download

setup:
	go get gopkg.in/spacemonkeygo/dbx.v1 && go mod vendor

generate: database/schema.dbx.go
	go generate ./database

test-no-generate:
	go test -count=1 ./...

test: generate
	make test-no-generate

run: generate test server
	./demoapi --loglevel debug --config config.hcl

runsqlite3-no-generate: test-no-generate server-no-generate
	./demoapi --loglevel debug --config config.hcl \
		--db_url "sqlite3:demoapi.sqlite3.db?sslmode=disable"

runsqlite3: generate test server
	./demoapi --loglevel debug --config config.hcl \
		--db_url "sqlite3:demoapi.sqlite3.db?sslmode=disable"

clean:
	rm -f ./demoapi demoapi.sqlite3.db*
	go clean
