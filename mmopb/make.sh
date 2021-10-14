#!/bin/bash

protoc -I . -I .. --go_out=plugins=grpc,paths=source_relative:. *.proto
