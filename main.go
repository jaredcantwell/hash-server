package main

import (
	"flag"
	"hash-server/server"
)

var flagPort int

func init() {
	flag.IntVar(&flagPort, "port", 8080, "port number on which to start listening for REST requests")
}

func main() {
	flag.Parse()
	server.New(flagPort).Run()
}
