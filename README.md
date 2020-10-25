# Rcon
[![golangci-lint](https://github.com/gorcon/rcon/workflows/golangci-lint/badge.svg)](https://github.com/gorcon/rcon/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/gorcon/rcon)](https://goreportcard.com/report/github.com/gorcon/rcon)
[![Build Status](https://travis-ci.org/gorcon/rcon.svg?branch=master)](https://travis-ci.org/gorcon/rcon)
[![Coverage](https://gocover.io/_badge/github.com/gorcon/rcon?0 "coverage")](https://gocover.io/github.com/gorcon/rcon)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/gorcon/rcon)

Source RCON Protocol implementation in Go.

## Protocol Specifications

RCON Protocol is described in the [valve documentation](https://developer.valvesoftware.com/wiki/Source_RCON_Protocol)

## Supported Games

* [Project Zomboid](https://store.steampowered.com/app/108600) 
* [Conan Exiles](https://store.steampowered.com/app/440900)
* [Rust](https://store.steampowered.com/app/252490) (add +rcon.web 0 to the args when starting the server)
* [ARK: Survival Evolved](https://store.steampowered.com/app/346110)

Open pull request if you have successfully used a package with another game with rcon support and add it to the list.

## Install

```text
go get github.com/gorcon/rcon
```

See [Changelog](CHANGELOG.md) for release details.

## Usage

```go
package main

import (
	"log"
	"fmt"

	"github.com/gorcon/rcon"
)

func main() {
	conn, err := rcon.Dial("127.0.0.1:16260", "password")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	response, err := conn.Execute("help")
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Println(response)	
}
```

## Requirements

Go 1.15 or higher

## Contribute

Contributions are more than welcome! 

If you think that you have found a bug, create an issue and publish the minimum amount of code triggering the bug so 
it can be reproduced.

If you want to fix the bug then you can create a pull request. If possible, write a test that will cover this bug.

## License

MIT License, see [LICENSE](LICENSE)
