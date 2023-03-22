zabbix [![GoDoc](https://godoc.org/github.com/AlekSi/zabbix?status.svg)](https://godoc.org/github.com/AlekSi/zabbix) [![Build Status](https://travis-ci.org/AlekSi/zabbix.svg?branch=master)](https://travis-ci.org/AlekSi/zabbix??branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/AlekSi/zabbix)](https://goreportcard.com/report/github.com/AlekSi/zabbix)
======

This Go package provides access to Zabbix API. 

| Zabbix Version | Compatibility |
|:---------------| :------------ |
| 6.4.x          | ✅            |
| 6.2.x          | ✅            |
| 6.0.x LTS      | ✅            |
| 5.4.x          | ✅            |
| 5.2.x          | ✅            |
| 5.0.x LTS      | ✅            |
| 4.4.x          | ✅            |
| 4.2.x          | ✅            |
| 4.0.x LTS      | ✅            |
| 3.4.x          | ✅            |
| 3.2.x          | ✅            |

Install it: `go get github.com/canghai908/zabbix-go`

You *have* to run tests before using this package – Zabbix API doesn't match documentation in few details, which are changing in patch releases. Tests are not expected to be destructive, but you are advised to run them against not-production instance or at least make a backup.

    export TEST_ZABBIX_URL=http://localhost:8080/zabbix/api_jsonrpc.php
    export TEST_ZABBIX_USER=Admin
    export TEST_ZABBIX_PASSWORD=zabbix
    export TEST_ZABBIX_VERBOSE=1
    go test -v

`TEST_ZABBIX_URL` may contain HTTP basic auth username and password: `http://username:password@host/api_jsonrpc.php`. Also, in some setups URL should be like `http://host/zabbix/api_jsonrpc.php`.

Documentation is available on [pkg.go.dev](https://pkg.go.dev/github.com/canghai908/zabbix-go).
Also, Rafael Fernandes dos Santos wrote a [great article](http://www.sourcecode.net.br/2014/02/zabbix-api-with-golang.html) about using and extending this package.

License: Simplified BSD License (see LICENSE).
