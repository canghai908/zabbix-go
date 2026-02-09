# zabbix [![GoDoc](https://godoc.org/github.com/AlekSi/zabbix?status.svg)](https://godoc.org/github.com/AlekSi/zabbix) [![Build Status](https://travis-ci.org/AlekSi/zabbix.svg?branch=master)](https://travis-ci.org/AlekSi/zabbix??branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/AlekSi/zabbix)](https://goreportcard.com/report/github.com/AlekSi/zabbix)

This Go package provides access to Zabbix API, Zabbix Sender Protocol, and Zabbix Get Protocol.

## Features

-   **Zabbix API**: Full support for Zabbix JSON-RPC API with automatic version detection
-   **Zabbix Sender Protocol**: Send monitoring data to Zabbix Server
-   **Zabbix Get Protocol**: Retrieve data from Zabbix Agent

## Version Compatibility

| Zabbix Version | Compatibility | Authentication Method |
| :------------- | :------------ | :-------------------- |
| 7.2+           | ✅            | Bearer Token (Header) |
| 7.2.x          | ✅            | JSON Body Auth        |
| 7.0.x LTS      | ✅            | JSON Body Auth        |
| 6.4.x          | ✅            | JSON Body Auth        |
| 6.2.x          | ✅            | JSON Body Auth        |
| 6.0.x LTS      | ✅            | JSON Body Auth        |
| 5.4.x          | ✅            | JSON Body Auth        |
| 5.2.x          | ✅            | JSON Body Auth        |
| 5.0.x LTS      | ✅            | JSON Body Auth        |
| 4.4.x          | ✅            | JSON Body Auth        |
| 4.2.x          | ✅            | JSON Body Auth        |
| 4.0.x LTS      | ✅            | JSON Body Auth        |
| 3.4.x          | ✅            | JSON Body Auth        |
| 3.2.x          | ✅            | JSON Body Auth        |

### Authentication Changes

Starting from **Zabbix 7.4**, the authentication token is sent in the HTTP `Authorization: Bearer` header instead of the JSON request body. This library automatically detects the Zabbix version and uses the appropriate authentication method. The version detection happens automatically when you call `Login()` or `SetAuth()`.

Install it: `go get github.com/canghai908/zabbix-go`

You _have_ to run tests before using this package – Zabbix API doesn't match documentation in few details, which are changing in patch releases. Tests are not expected to be destructive, but you are advised to run them against not-production instance or at least make a backup.

    export TEST_ZABBIX_URL=http://localhost:8080/zabbix/api_jsonrpc.php
    export TEST_ZABBIX_USER=Admin
    export TEST_ZABBIX_PASSWORD=zabbix
    export TEST_ZABBIX_VERBOSE=1
    go test -v

`TEST_ZABBIX_URL` may contain HTTP basic auth username and password: `http://username:password@host/api_jsonrpc.php`. Also, in some setups URL should be like `http://host/zabbix/api_jsonrpc.php`.

For integration tests of Sender and Get protocols, set:

-   `TEST_ZABBIX_SERVER`: Zabbix Server address (e.g., `localhost:10051`)
-   `TEST_ZABBIX_AGENT`: Zabbix Agent address (e.g., `localhost:10050`)

## Usage Examples

### Zabbix API

```go
package main

import (
    "github.com/canghai908/zabbix-go"
)

func main() {
    // Create API client
    api := zabbix.NewAPI("http://localhost/zabbix/api_jsonrpc.php")

    // Login (automatically detects version and uses appropriate auth method)
    auth, err := api.Login("Admin", "zabbix")
    if err != nil {
        panic(err)
    }

    // Use API
    hosts, err := api.HostsGet(zabbix.Params{})
    if err != nil {
        panic(err)
    }

    // Access version info
    versionInfo := api.GetVersionInfo()
    // versionInfo contains: Version, Major, Minor, Patch, UseBearer, UseUsername
}
```

### Zabbix Sender Protocol

```go
package main

import (
    "github.com/canghai908/zabbix-go"
    "time"
)

func main() {
    // Create sender
    sender := zabbix.NewSender("localhost", 10051)
    sender.SetTimeout(10 * time.Second)

    // Send single data item
    data := zabbix.SenderData{
        Host:  "test-host",
        Key:   "test.key",
        Value: "123",
        Clock: time.Now().Unix(), // Optional, 0 means current time
    }

    response, err := sender.Send(data)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Response: %s, Info: %s\n", response.Response, response.Info)

    // Send multiple data items in batch
    batch := []zabbix.SenderData{
        {Host: "host1", Key: "key1", Value: "value1"},
        {Host: "host2", Key: "key2", Value: "value2"},
    }

    response, err = sender.SendBatch(batch)
    if err != nil {
        panic(err)
    }
}
```

### Zabbix Get Protocol

```go
package main

import (
    "github.com/canghai908/zabbix-go"
    "time"
)

func main() {
    // Create get client
    get := zabbix.NewGet("localhost", 10050)
    get.SetTimeout(10 * time.Second)

    // Get single value
    value, err := get.GetValue("system.uptime")
    if err != nil {
        panic(err)
    }

    fmt.Printf("Uptime: %s\n", value)

    // Get multiple values
    keys := []string{"system.uptime", "system.hostname", "system.cpu.load"}
    values, err := get.GetValues(keys)
    if err != nil {
        panic(err)
    }

    for key, value := range values {
        fmt.Printf("%s: %s\n", key, value)
    }
}
```

## Protocol Details

### Zabbix Sender Protocol

The Zabbix Sender Protocol is used to send monitoring data to Zabbix Server. The protocol uses TCP connections and a binary format:

-   **Header**: `ZBXD\x01` (5 bytes)
-   **Data Length**: 8 bytes (little-endian)
-   **JSON Data**: Array of objects with `host`, `key`, `value`, and optional `clock` fields

Default port: **10051**

### Zabbix Get Protocol

The Zabbix Get Protocol is used to retrieve data from Zabbix Agent. The protocol uses simple text-based communication:

-   Send: `key\n`
-   Receive: `value\n` or error message

Default port: **10050**

## Documentation

Documentation is available on [pkg.go.dev](https://pkg.go.dev/github.com/canghai908/zabbix-go).

Also, Rafael Fernandes dos Santos wrote a [great article](http://www.sourcecode.net.br/2014/02/zabbix-api-with-golang.html) about using and extending this package.

## License

Simplified BSD License (see LICENSE).
