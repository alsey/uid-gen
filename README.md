# Global Sequence Unique ID Generator

Generate global unique id in sequence.

The concept is 2-layer generating system, first layer generates unique id by redis service, and second layer persistences id by mysql backend.

When the server start, it will synchronizes MySQL and Redis first, so it can be used in distribute system envionment safely.

## Installation

```bash
$ go get github.com/alsey/uid-gen
```

## Connfiguration

  Use envionment variables to config this service.

- PORT0 : port, default is 3000
- MYSQL_DSN : mysql connection string, default is root:password@tcp(127.0.0.1:3306)/counter?charset=utf8
- REDIS_HOST : redis service address, default is 127.0.0.1
- REDIS_PORT : redis service port, default is 6379

## How to Use

1. Use a key, generate unique id in sequence

Every time access the URL, the server will return a unique id. Default the id starts with 1, then 2, 3, 4... in sequence.

```
http://<host>:<port>/<some_key>
```

2. Set the start number

You can set the start number not the default 1.

```
http://<host>:<port>/<some_key>?start=99
```

The key starts with 99, then 100, 101, 102... in sequence.

3. Set the step

You can set the step not the default 1.

```
http://<host>:<port>/<some_key>?step=2
```

The key step is 2, first time return 1, then 3, 5, 7, 9... in sequence.

The step also can be negative number. 

```
http://<host>:<port>/<some_key>?start=100&step=-2
```

The service return 100, then 98, 96, 94... in sequence.

4. I wrote a Dockerfile for you, you can make a Docker image, and put it in Kubernetes or Mesos.

```bash
$ docker build -t 'mongo-image-server' .
```

5. The server includes a health check endpoint for microservice scenario.

```
http://<host>:<port>/health
http://<host>:<port>/env
```

## License

  [MIT](LICENSE)