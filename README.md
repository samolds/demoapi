# DemoAPI

A containerized RESTful HTTP microservice that stores user, group, and
membership data.


## About

This is a small RESTful HTTP backend, written in Go, that manages simple
`user`, `group`, and `membership` records. A `user` can belong to multiple
`groups`, and a `group` can have multiple `users`. The many-to-many
relationship is tracked by the `membership` record.


## Dependencies

- Docker

OR

- Go 1.13+
- Sqlite3


## How To: with Docker

```sh
make up
curl http://localhost:8080/
```

To cleanup local images:

```
make clean-docker
```


## How To: from src (without Prometheus and Grafana servers)

```sh
cd go/src/demoapi
make runsqlite3-no-generate
curl http://localhost:8080/
```


## Dummy Data

There is also a helper python script that will POST up a number of users and
groups and create memberships so that there is dummy data available to poke.

NOTE: The python script depends on the "urllib.requests" and "json" libraries.

```sh
make up
python post_dummy_data.py
```


### Example cURLs

- Create a new user
```sh
curl -X POST --data-binary '{"first_name": "Matt", "last_name": "F", "userid": "mattf", "groups": ["nasa"]}' http://localhost:8080/users
```

- Get user data
```sh
curl http://localhost:8080/users/{userid}
```

- Update a specific user's memberships
```sh
curl -X PUT --data-binary '{"groups": ["group1", "group2"]}' http://localhost:8080/users/user1
```

- Delete user and their memberships
```sh
curl -X DELETE http://localhost:8080/users/{userid}
```

- Get all possible users, paged
```sh
curl http://localhost:8080/users?quantity=10&offset=10
```

- Create a new group
```sh
curl -X POST --data-binary '{"name": "group1"}' http://localhost:8080/groups
```

- Get group data
```sh
curl http://localhost:8080/groups/{groupname}
```

- Update a specific group's memberships
```sh
curl -X PUT --data-binary '{"userids": ["user1", "user2"]}' http://localhost:8080/groups/group1
```

- Delete group and its memberships
```sh
curl -X DELETE http://localhost:8080/groups/{groupname}
```

- Get all possible groups, paged
```sh
curl http://localhost:8080/groups?quantity=10&offset=10
```

If you have jq installed:
```sh
curl -s http://localhost:8080/users?limit=10&token=10 | jq
```


### Additional features not in the spec

- `GET /users`
  Will return paginated user objects with links to the next page
- `GET /groups`
  Will return paginated group objects with links to the next page
- `GET /metrics`
  Will return Prometheus metrics that can be used by the Grafana server,
  visible at [localhost:3000](http://localhost:3000). (See note below about
  connecting the Prometheus server as a Grafana datasource)
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


### Using Grafana

```sh
make up
```

This will spin off the main API server with Prometheus metric collection, as
well as a Prometheus proxy server and a Grafana server. Once everything is up
and running, go to [localhost:3000](http://localhost:3000). You should be
greeted by the Grafana log in page. On first login, the creds are `admin` and
`admin`.

Once you're in, go to the [datasources](http://localhost:3000/datasources) tab,
select "Prometheus" and enter `http://127.0.0.1:9090` as the "URL" and change
"Access" to `Browser`. Then scroll down and test the connection.

At this point, [Create a New Dashboard](http://localhost:3000/dashboard/new)
and "Add a new panel". Once here, you should be able to start visualizing
metrics collected for the API server.
