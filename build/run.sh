#!/bin/bash

# Компиляция приложения
go build -o agent ../cmd/agent/main.go

for i in {1..100}
do
    ./agent &
    sleep 3.5
done