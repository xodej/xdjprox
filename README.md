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

## Installation

Compile with go for your desired platform and run the xdjprox binary.

```cli
# running xdjprox with defaults
./xdjprox
```

```cli
# -h help
./xdjprox -h

Usage of xdjprox:
  -i string
        xdjprox port (default ":8080")
  -o string
        Jedox OLAP (default "http://127.0.0.1:7777")
  -req
        log http request (default false)
  -res
        log OLAP http response (default false)
  -w    
        enable write requests (default false)
```

```cli
# running xdjprox overriding defaults
# -i   <PROXY port>
# -o   <OLAP address>
# -req log request
# -res log response
# -w   enable OLAP writes
./xdjprox -i :8080 -o http://olap.example.org:7777 -req -res -w
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
