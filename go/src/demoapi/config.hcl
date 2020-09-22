//db_url = "postgres://dbuser:dbpass@db/demoapi?sslmode=disable"
//db_url = "sqlite3:demoapi.sqlite3.db?sslmode=disable"

api_slug    = "demoapi"
api_addr    = ":8080"
metric_addr = ":9090"

graceful_shutdown_timeout_sec = 5
write_timeout_sec             = 15
read_timeout_sec              = 15
idle_timeout_sec              = 15

loglevel = "debug"
developer_mode = true
insecure_requests_mode = true
