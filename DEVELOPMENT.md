# Development environment setup, building, running, and testing

## Prerequisites

Configure your environment file:

```
cp .env.template .env
```

Specify database connection parameters, front-end URL (necessary for the "Access-Control-Allow-Origin" HTTP header) and other variables if needed.

# Build

## Uncontainerized

`go build -o ./build/main ./cmd/server/main.go`

To build and run the service in a Docker container, see the *Run* section below.


# Run


## All the services using docker-compose

Specify the GOMODCACHE environment variable if you prefer downloading dependencies to a directory different from the default one (to save building time and network traffic, the GOMODCACHE directory of the `log_relay` container is mounted to the GOMODCACHE directory of the host.)

```
[GOMODCACHE=`pwd`] docker compose up
```


## Database

`docker compose up database`


## Service

```
[GOMODCACHE=`pwd`] docker compose up log_relay
```

or run the service uncontainerized:

`./build/main`


# Test

## Automated

[TBD - unit and E2E tests]

## Manual

### Register a user account (save the Set-Cookie headers to a file)

`curl -iv -X POST -c lr_cookie.txt http://127.0.0.1:8080/api/user/register -d '{"name":"user1","email":"email@example.com","password":"Ab_12345678"}'`

Output excerpt:

```
...
Set-Cookie: token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImVtYWlsQGV4YW1wbGUuY29tIiwiZXhwIjoxNzc1MjY2NTM5LCJpYXQiOjE3NzUxODAxMzksInVzZXJfaWQiOjF9.ZPl7kIIR6Rc-zIhiYdfG1AI8CEiZibpopeBEfQ8Bg4k; Path=/; Max-Age=86400; HttpOnly
...
```

`curl -iv -X GET -b lr_cookie.txt 'http://127.0.0.1:8080/api/list/index?cnt_per_page=10&page=1'`


### Log in

`curl -iv -X POST -c lr_cookie.txt` http://127.0.0.1:8080/api/user/login -d '{"email":"email@example.com","password":"Ab_12345678"}'`


```
200 OK
...
Set-Cookie: token=...
...
{"is_admin":false,"message":"Login successful","user_email":"email@example.com","user_name":"user1"}
```


### Create lists

`curl -iv -X POST http://127.0.0.1:8080/api/list/create -b lr_cookie.txt -d '{"name":"Common Mailing","listtype":"MAILING","public_facing_name":"Common Mailing List"}'`


```
201 Created
...
{"id":1,"created_at":"2026-04-02T22:27:42.572847345-05:00","updated_at":"2026-04-02T22:27:42.572847345-05:00","PublicID":"KHU7UUl9X1x16DGlBB4kGm5xWDXVfcp-BOYwaJZjdzs","public_facing_name":"Common Mailing List","name":"Common Mailing","list_type":"MAILING"}
```

`curl -iv -X POST http://127.0.0.1:8080/api/list/create -b lr_cookie.txt -d '{"name":"Common Inquiry","listtype":"INQUIRY","public_facing_name":"Common Inquiry List"}'`

```
201 Created
...
{"id":2,"created_at":"2026-04-06T03:25:29.488961768Z","updated_at":"2026-04-06T03:25:29.488961768Z","PublicID":"l9w15PPiwP7yMm4h22lqR1N5nNg2OfVhimNFdTerZFU","public_facing_name":"Common Inquiry List","list_type":"INQUIRY","name":"Common Inquiry","subscriber_count":0}
```


### Get the list index

`curl -iv -X GET -b lr_cookie.txt 'http://127.0.0.1:8080/api/list/index?cnt_per_page=10&page=1'`

```
200 OK
...
{"current_page":1,"lists":[{"id":2,"created_at":"0001-01-01T00:00:00Z","updated_at":"2026-04-06T03:25:29.488961Z","PublicID":"","public_facing_name":"","list_type":"INQUIRY","name":"Common Inquiry","subscriber_count":0},{"id":1,"created_at":"0001-01-01T00:00:00Z","updated_at":"2026-04-03T03:27:42.572847Z","PublicID":"","public_facing_name":"","list_type":"MAILING","name":"Common Mailing","subscriber_count":0}],"total_count":2,"total_pages":1}
```

