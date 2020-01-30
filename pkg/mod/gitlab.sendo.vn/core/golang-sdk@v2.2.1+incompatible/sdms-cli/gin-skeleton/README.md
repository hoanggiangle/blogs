Service {{ .ServiceName }}
===============

## Setup

View generic document at [SDK Golang](https://gitlab.sendo.vn/core/golang-sdk/wikis/home)

## Run

#### Use docker-compose

- Source can be put anywhere, it will be mount to GOPATH when docker start

    ``` bash
    dep ensure # or go get
    docker-compose up myapp

    ###### some useful commands
    docker-compose restart myapp
    docker-compose logs -f myapp
    docker-compose exec myapp bash
    ```
- Or enter bash shell

    ``` bash
    docker-compose run --rm -p8000:8000 myapp bash
    > go run main.go run
    ```

##### If need override config

- Create a file named `docker-compose.override.yaml` like:

    ``` yaml
    version: "3.2"

    services:
      myapp:
        environment:
          PORT: 1234
        ports:
        - 8000:1234
    ```

#### Manual
- Install mongo, redis, ...
- Setup `GOPATH`
- Run:

    ``` bash
    dep ensure # or go get
    go run main.go run

    #### to create config file
    go run main.go outenv > .env
    ```
