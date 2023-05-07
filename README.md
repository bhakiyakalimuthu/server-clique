### ***SERVER-CLIQUE (server-client-queue)***
* This repo containing code for a server and clients that communicate via message queue.
* Server has data structure that holds the data in the memory while keeping the order of items as they added.
* Server reads the request events(client request) from message queue and act accordingly.
* Client request server to AddItem(key, value), RemoveItem(key), GetItem(key), GetAllItems() via message queue. 
* Clients can be added / removed while not interfering to the server or other clients.
* `RabbitMQ` is used as a message queue.

# Prerequisites
- Go 1.19
- Ubuntu 20.04 (any linux based distros) / OSX

# Build & Run
* Application can be build and started by using Makefile.
* Make sure to cd to project folder.
* Run the below commands in the terminal shell.
* Make sure to run `Pre-run` and `Go path` is set properly.

# Pre-run
    make mod
    make lint
    make clean


# How to run build
    make build

# Setup
* Client actions are configured via json file which is part of `server-clique/client/input.json`.
* Client can perform actions such as
  * AddItem(key, value)
  * RemoveItem(key)
  * GetItem(key) 
  * GetAllItems()
* Each row in the json array represents action.
* Action require 3 values such as `action, key, value`.
     ```json
     {"action": "add","key": "O","value": "o"},
     {"action": "getall"},
     {"action": "get","key": "O"},
     {"action": "remove","key": "O"},
     ```
* Server outputs successful response to `server-clique/output.json` file.
* `make clean` can cleanup `output.json` file. 

# How to run
> There are 3 ways server,client & queue can be run.
> Make sure to complete above Setup 

> **Note:**
>* Make sure you have docker installed or Download and install rabbitmq installer.
>* Creating docker image for both client and server has different docker run commands for different platforms amd64(debian/x86-64) & arm64(osx M1/M2).

## I.Start via main file
### Step:1 start rabbitmq
    docker run --rm -it -p 15672:15672 -p 5672:5672 rabbitmq:3-management

### Step:2 start server
    go run -race cmd/server/main.go OR go run -race cmd/server/mem_optimised/main.go

### Step:3 start client
    go run -race cmd/clien/main.go

## II.Start via locally build binary
Run these commands in terminal shell

* `cd server-clique`
* `make build`
* `./server-clique-server`
* `./server-clique-client`
* `make clean` to cleanup locally built binary as well as output.json file.

## III.Run as docker container
### Step:1
     make docker-run-rabbitmq
### Step:2
    make docker-image-server / make osx-docker-image-server 
    make docker-run-server
### Step:3
    make docker-image-client / make osx-docker-image-client
    make docker-run-client

# Testing
## How to run unit test
    make test
## How to run performance test
### Step:1
     make docker-run-rabbitmq
### Step:2
    go run -race cmd/server/main.go OR go run -race cmd/server/mem_optimised/main.go
### Step:3
    sh client_loop.sh 

>* Script to fire up messages from client.
>* This simple shell script starts client in loop and messages are feed from input.json.
>* Multiple client simulation: Same script can be run as multiple client instance to increase the load.
>* Any client can be used But only requirement would be to follow the message type defined in `types` package `types.Message`.

# Memory store benchmark
* There are 2 memory store implementation available. 
  * `memstore.go` and `memstore_optimised.go` is part of server
  * benchmark given below for both implementation
```shell
❯ go test -bench=. -benchmem
goos: darwin
goarch: arm64
pkg: github.com/bhakiyakalimuthu/server-clique/server
BenchmarkServer_Memstore-10                       155757              7738 ns/op            5691 B/op          5 allocs/op
BenchmarkServer_MemstoreOptimised-10              172684              6900 ns/op            4939 B/op          4 allocs/op
PASS
ok      github.com/bhakiyakalimuthu/server-clique/server        7.730s

```
# Sample output
```
2023/05/07 08:46:40.819582 worker id:4 performed action:add key:A value:a
2023/05/07 08:46:40.819731 worker id:1 performed action:add key:B value:b
2023/05/07 08:46:40.819747 worker id:5 performed action:add key:C value:c
2023/05/07 08:46:40.819822 worker id:3 performed action:get key:A value:a
2023/05/07 08:46:40.819871 worker id:2 performed action:remove key:B
2023/05/07 08:46:40.819915 worker id:1 performed action:add key:B value:b
2023/05/07 08:46:40.819969 worker id:4 performed action:getall items:[{A a 1683449200812541000} {B b 1683449200812859000} {C c 1683449200812920000}] itemsLength:3
2023/05/07 08:46:40.820010 worker id:5 performed action:add key:D value:d
2023/05/07 08:46:40.820164 worker id:3 performed action:add key:E value:e
```
> ***Improvements***
>* Retry if rabbitMQ abruptly close the connection.
>* Quit server if rabbitMQ connection is closed as there is no retry logic. 
