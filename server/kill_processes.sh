#!/bin/bash

# Find and kill all running server processes
kill -9 $(lsof -ti:5001,5002,5003,5004) &
echo "All active servers have been killed! - ctrl+c to exit"
