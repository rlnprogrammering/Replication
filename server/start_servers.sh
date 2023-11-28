#!/bin/bash

# Default value for the duration
default_minutes=5

# Function to display usage information
usage() {
    echo "Usage: $0 [-min <minutes>]"
    exit 1
}

# Parse command-line options
while [ "$#" -gt 0 ]; do
    case "$1" in
        -min)
            if [ -n "$2" ] && [ "$2" -eq "$2" ] 2>/dev/null; then
                minutes="$2"
                shift 2
            else
                echo "Error: Invalid value for -min."
                usage
            fi
            ;;
        *)
            echo "Error: Unknown option: $1"
            usage
            ;;
    esac
done

# If -min option is not provided, use the default value
if [ -z "$minutes" ]; then
    minutes="$default_minutes"
fi

# Run server.go with the specified minutes
for ((i=0; i<4; i++)); do
    go run server.go -min "$minutes" &
done

echo "All servers started with a duration of $minutes minutes!"
