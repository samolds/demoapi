# DemoAPI

A containerized RESTful HTTP microservice that stores user, group, and
membership data.


### Dependencies

- Docker

OR

- Go 1.10+
- Sqlite3


### How To with Docker

```sh
make up
curl http://localhost:8080/
```

There is also a helper python script that will POST up a number of users and
groups and create memberships so that there is dummy data available to poke.

NOTE: The python script depends on the "urllib.requests" and "json" libraries.

```sh
make up
python post_dummy_data.py
```


### How To from src without Prometheus and Grafana servers

```sh
cd go/src/demoapi
make runsqlite3-no-generate
curl http://localhost:8080/
```


### Additional features not in the spec

- `GET /users`
  Will return paginated user objects with links to the next page
- `GET /groups`
  Will return paginated group objects with links to the next page
- An `insecure_requests_mode` flag that can be set in go/src/demoapi/config.hcl
  that toggles the need for an Authorization header with each request. For now,
  the token just needs to be any non-empty string.
- Support for either sqlite or postgres depending on the provided database
  configuration variables
- The entire project is containerized and stood up with docker-compose.

If the `insecure_requests_mode = false` configuration is set in config.hcl,
then an Authorization token must be provided with each request, like:

```
curl -H "Authorization: Bearer dummy_token" http://localhost:8080/groups
```


### More example cURLs

```sh
curl http://localhost:8080/groups

curl http://localhost:8080/groups?quantity=10&offset=10

curl http://localhost:8080/users/{group_id}

curl http://localhost:8080/users/{group_id}?quantity=10&offset=10

curl -X POST --data-binary '{"first_name": "Matt", "last_name": "F", "userid": "mattf", "groups": ["nasa"]}' http://localhost:8080/users
```

If you have jq installed:

```sh
curl -s http://localhost:8080/users?limit=10&token=10 | jq
```


### TODOs

- grafana/prometheus
