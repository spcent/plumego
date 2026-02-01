#!/bin/bash
cd "$(dirname "$0")"
echo "Building CRUD Demo..."
go build -o crud-demo main_simple.go
if [ $? -eq 0 ]; then
    echo "Starting server..."
    ./crud-demo
else
    echo "Build failed"
fi
