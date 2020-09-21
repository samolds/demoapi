# DemoAPI Server

This is a backend REST server to manage user and group memberships.


### Running Locally

```sh
make run
```


### Gotchas

- If you would like to regenerate database/schema.dbx.go, you will probably
  need an existing build of DBX. It appears that dbx is no longer `go get`-able
  due to missing dependencies...
