#!/bin/bash

# Компиляция приложения
go build -o agent ../cmd/agent/main.go

for i in {1..5}
do
    count=$(( ( RANDOM % i ) + 1 ))
    ./agent -count $count -interval 0.5s -timeout 10s --hostname agent$count
    sleep 0.5
done