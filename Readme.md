**Note: All scripts are heavily inspired from ChatGPT.**

## Start/close servers

Navigate to server folder and run:

**Start a single server | maximum of 4 servers**

`go run server.go -min <minutes>` | minutes flag is optional, default 5 minutes

**Run 4 servers at once**

`./start_servers.sh -min <minutes>` | minutes flag is optional, default 5 minutes

**Close all processes (servers & clients)**

`./kill_processes.sh` 

**Fail a single server**

`./fail_server.sh <port>` | e.g. `./fail_server.sh 5001`

## Start/use client
Navigate to client folder and run:

`go run client.go -id <id(int)>`

- Bid: Enter your bid integer value
   - Service will respond with status: `success` or `fail`
- Result: Type `result` in client terminal
   - Will return highest current bid across all clients

 


