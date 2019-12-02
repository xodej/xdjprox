# xdjprox

## a proxy written in go for the Jedox Excel Add-In

This repository is **unstable** and **not for production use**.

The intention is to show a proof-of-concept for Jedox Excel Add-In read-only connections.
All potentially harmful HTTP requests to the OLAP are filtered before being passed to the OLAP;
independent from user rights. This allows users to work safely with Excel while still being
able to write back to the OLAP (e.g. planning via Web-Reports) or a
dedicated second Excel connection.

Please note that enabling the Jedox Excel modeler bypasses restrictions by calling actions
directly in the Jedox web frontend which is not affected by the proxy.

The proxy will listen on port :8080 by default and redirect all whitelisted OLAP calls to
port :7777 which is not affected and still works as before. To make use of the proxy
you have to change the port of the connection from :7777 to :8080.

**By default login credentials are logged and visible.**

## Installation

Compile with go for your desired platform and run the xdjprox binary.

```cli
# running xdjprox with defaults
./xdjprox
```

```cli
# show help
./xdjprox -h

Usage of xdjprox:
  -i string
        xdjprox local port (default ":8080")
  -log-all
        enable logging everything (default false)
  -log-file string
        log file name
  -log-req
        enable logging of client http request (default false)
  -log-res
        enable logging of OLAP http response (default false)
  -o string
        Jedox OLAP server address (default "http://127.0.0.1:7777")
  -w    enable write requests (default false)
```

```cli
# running xdjprox overriding defaults
# -i        <PROXY port>
# -log-all  Log request and response
# -log-file Log to file
# -log-req  Log request
# -log-res  Log response
# -o        <OLAP address>
# -w        enable OLAP writes
./xdjprox -i :8080 -o http://olap.example.org:7777 -log-all -w -log-file res_req.log
```

## Output

```json
{"level":"info","msg":"xdjprox started with config \u0026main.Config{TargetURL:\"http://127.0.0.1:7777\", EntryURL:\":8080\", TimeFormat:\"2006-02-01 15:04:05\", LogRequest:true, LogResponse:true, EnableWrite:true, LogFile:\"res_req.log\", LogAll:false}","time":"2019-12-02T20:49:01+01:00"}
{"level":"info","msg":"forwarded request /server/info","request_id":"5a9effca-07ba-4601-adf4-f67cd1be39fd","session":"","time":"2019-12-02T20:49:13+01:00","type":"forward"}
{"level":"info","msg":"GET /server/info HTTP/1.1\r\nHost: 127.0.0.1:7777\r\nAccept-Charset: utf-8\r\nAccept-Encoding: gzip, deflate\r\nConnection: keep-alive\r\nContent-Length: 0\r\nContent-Type: text/plain\r\nX-Forwarded-Host: \r\nX-Palo-Sv: 0\r\n\r\n","request_id":"5a9effca-07ba-4601-adf4-f67cd1be39fd","session":"","time":"2019-12-02T20:49:13+01:00","type":"request"}
{"level":"info","msg":"HTTP/1.1 200 OK\r\nContent-Length: 33\r\nContent-Type: text/plain;charset=utf-8\r\nServer: Palo\r\nX-Palo-Sv: 1639917714\r\n\r\n19;3;5;10339;0;0;1804290315;0;D;\n","request_id":"5a9effca-07ba-4601-adf4-f67cd1be39fd","session":"","time":"2019-12-02T20:49:13+01:00","type":"response"}
{"level":"info","msg":"forwarded request /server/login","request_id":"a61a1b0b-a832-4d34-99dd-060545b40650","session":"","time":"2019-12-02T20:49:13+01:00","type":"forward"}
{"level":"info","msg":"POST /server/login HTTP/1.1\r\nHost: 127.0.0.1:7777\r\nAccept-Charset: utf-8\r\nAccept-Encoding: gzip, deflate\r\nConnection: keep-alive\r\nContent-Length: 274\r\nContent-Type: text/plain\r\nX-Forwarded-Host: \r\nX-Palo-Sv: 0\r\n\r\nuser=admin\u0026extern_password=admin\u0026machine=127.0.0.1\u0026required=\u0026optional=3a00c07f02\u0026new_name={%22client%22:%22Excel%20Add-in%22,%22client_ver%22:%2216866%22,%22lib%22:%22libpalo_ng%22,%22lib_ver%22:%2219.3.2.5947%22,%22desc%22:%22user%20login%22}\u0026external_identifier=de_DE","request_id":"a61a1b0b-a832-4d34-99dd-060545b40650","session":"","time":"2019-12-02T20:49:13+01:00","type":"request"}
{"level":"info","msg":"HTTP/1.1 200 OK\r\nContent-Length: 46\r\nContent-Type: text/plain;charset=utf-8\r\nServer: Palo\r\nX-Palo-Sv: 1639917715\r\n\r\nB9i0QF5A5xhFQvvTFkRcRz1dp4TbQqMs;300;3b;0;\"\";\n","request_id":"a61a1b0b-a832-4d34-99dd-060545b40650","session":"","time":"2019-12-02T20:49:13+01:00","type":"response"}
{"level":"info","msg":"forwarded request /server/logout","request_id":"c011b5cb-ddc5-4821-aa6f-bfe13419a5c9","session":"B9i0QF5A5xhFQvvTFkRcRz1dp4TbQqMs","time":"2019-12-02T20:49:13+01:00","type":"forward"}
{"level":"info","msg":"GET /server/logout?sid=B9i0QF5A5xhFQvvTFkRcRz1dp4TbQqMs HTTP/1.1\r\nHost: 127.0.0.1:7777\r\nAccept-Charset: utf-8\r\nAccept-Encoding: gzip, deflate\r\nConnection: keep-alive\r\nContent-Length: 0\r\nContent-Type: text/plain\r\nX-Forwarded-Host: \r\nX-Palo-Sv: 1639917714\r\n\r\n","request_id":"c011b5cb-ddc5-4821-aa6f-bfe13419a5c9","session":"B9i0QF5A5xhFQvvTFkRcRz1dp4TbQqMs","time":"2019-12-02T20:49:13+01:00","type":"request"}
{"level":"info","msg":"HTTP/1.1 200 OK\r\nContent-Length: 3\r\nContent-Type: text/plain;charset=utf-8\r\nServer: Palo\r\nX-Palo-Sv: 1639917716\r\n\r\n1;\n","request_id":"c011b5cb-ddc5-4821-aa6f-bfe13419a5c9","session":"B9i0QF5A5xhFQvvTFkRcRz1dp4TbQqMs","time":"2019-12-02T20:49:13+01:00","type":"response"}
```

## Documentation

For more examples please look [here](./docs/index.md).

## OLAP call whitelist
```
/server/change_password
/server/databases
/server/info
/server/licenses
/server/login
/server/logout
/server/user_info

/database/cubes
/database/dimensions
/database/info

/dimension/cubes
/dimension/dfilter
/dimension/element
/dimension/elements
/dimension/info

/element/info

/cube/holds
/cube/info
/cube/locks
/cube/rules

/cell/area
/cell/drillthrough
/cell/export

/cell/value
/cell/values

/rule/functions
/rule/info
/rule/parse

/svs/info

/view/calculate

/meta-sp

/api
/inc/
/favicon.ico
```


## License

Licensed under the [MIT](./LICENSE) license.
