package main

import (
	"os"
	"./client"
	"./server"
)

func main() {
	types := os.Args[1]
	switch types {
	case "server":
		server.Main()
	case "client":
		client.Main()

	}
}