### SERVER-CLIQUE (server-client-queue)
>* This repo containing code for a server and client that communicate via message queue.
>* Rabbit mq is used as message queue.
>* Client request server to AddItem(key, value), RemoveItem(key), GetItem(key), GetAllItems()
     via rabbitmq
>* Server has data structure that holds the data in the memory while keeping the order of items as they added.
>* Server reads the request events(client request) from rabbitmq and act accordingly.

# Prerequisites
- Go 1.19
- Ubuntu 20.04 (any linux based distros) / OSX

# Build & Run
* Application can be build and started by using Makefile.
* Make sure to cd to project folder.
* Run the below commands in the terminal shell.
* Make sure to run Pre-run and Go path is set properly

# Pre-run
    make mod
    make lint
    make clean


# How to run build
    make build

# Setup
* Client actions are configured via json file which is part of `server-clique/client/input.json`
* Client can perform actions such as
  * AddItem(key, value)
  * RemoveItem(key)
  * GetItem(key) 
  * GetAllItems()
* Each row in the json array represents action
* Action require 3 values such as `action, key, value`
     ```json
     {"action": "add","key": "O","value": "o"},
     {"action": "getall"},
     {"action": "get","key": "O"},
     {"action": "remove","key": "O"},
     ```
* Server outputs successful response to `server-clique/output.json` file
* `make clean` can cleanup `output.json` file 

# How to run
> There are 3 ways server,client & queue can be run.
> Make sure to complete above Setup 

> **Note:**
> Make sure you have docker installed or Download and install rabbitmq installer

## I.Start via main file
### Step:1 start rabbitmq
`docker run --rm -it -p 15672:15672 -p 5672:5672 rabbitmq:3-management`

### Step:2 start server
`go run cmd/server/main.go`

### Step:3 start client
`go run cmd/clien/main.go`

## II.Start via locally build binary
Run these commands in terminal shell

* `cd server-clique`
* `make build`
* `./server-clique-server`
* `./server-clique-client`
* `make clean` to cleanup locally built binary as well as output.json file

## III.Run as docker container
### Step:1
     make docker-run-rabbitmq
### Step:2
    make docker-image-server
    make docker-run-server
### Step:3
    make docker-image-client
    make docker-run-client

# Testing
## How to run unit test
    make test
## How to run performance test
### Step:1
     make docker-run-rabbitmq
### Step:2
    `go run cmd/server/main.go`
### Step:3
    sh client_loop.sh 

>* Script to fire up messages from client.
>* This simple shell script starts client in loop and messages are feed from input.json

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

> ***Improvements***
>* Retry if rabbitmq abruptly close the connection
>* Quit server if rabbit mq connection is closed as there is no retry logic 
