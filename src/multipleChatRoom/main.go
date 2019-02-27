package main

import (
	"os"
	"./client"
	"./server"
)

func main() {
	types := os.Args[1]
	switch types {
	case "s":
		server.Main()
	case "server":
		server.Main()
	case "c":
		client.Main()
	case "client":
		client.Main()
	}
}