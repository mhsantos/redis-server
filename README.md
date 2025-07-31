# redis-server

A Redis server implementing the [REST Protocol](https://redis.io/docs/latest/develop/reference/protocol-spec/).

## Usage

To run the server you can either execute the compiled version or run from source:

```
# running the compiled version
./bin/server

# running from source
go run server/main.go
```

## Compiling

To compile the source code into the bin directory:

```
go build -o bin/server ./server
```

## Supported commands

At the moment, the following commands are supported:

* DEL
* EXISTS
* GET
* SET
* INCR
* TTL

## Making requests to the server

The server accepts requests in the format defined by the RESP protocol.

You can either use the official redis-client implementation for this or make requests to the default port 6379.

Now an example to run the following commands:

```
*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$10\r\nredis-test\r\n
*3\r\n$3\r\nSET\r\n$5\r\nname2\r\n$11\r\nredis-test2\r\n
*2\r\n$3\r\nGET\r\n$4\r\nname\r\n$10\r\n
*2\r\n$3\r\nGET\r\n$5\r\nname2\r\n
*2\r\n$3\r\nDEL\r\n$5\r\nname2\r\n
*2\r\n$3\r\nGET\r\n$5\r\nname2\r\n
```

Using netcat:
```
$ nc localhost 6379 -c
*3
$3
SET
$4
name
$10
redis-test
+OK

*3
$3
SET
$5  
name2
$11
redis-test2
+OK

*2
$3 
GET
$4
name
$10
redis-test

*2
$3
GET
$5
name2
$11
redis-test2

*2
$3
DEL
$5
name2
:1

*2
$3
GET
$5
name2
+not found
```