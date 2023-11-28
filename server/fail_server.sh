#!/bin/bash

# Check if the port argument is provided
if [ -z "$1" ]; then
    echo "Usage: $0 <port>"
    exit 1
fi

port="$1"

# Get the PID of the server processes on the specified port
server_pids=$(lsof -i :$port | awk '/server/ {print $2}')

# Check if there are server processes to kill
if [ -z "$server_pids" ]; then
    echo "No server processes found on port $port."
    exit 0
fi
count=0
# Kill each server process
for pid in $server_pids
do
    count=$((count+1))
    kill $pid
done
((count -= 1))
echo "Server processes on port $port have been terminated. 1 server + $count client listeners have been killed."
