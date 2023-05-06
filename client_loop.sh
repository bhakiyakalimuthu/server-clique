#!/bin/bash

for ((i = 1; i <= 100; i++)); do
    go run -race cmd/client/main.go
done
