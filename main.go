package main

import (
	client "go_wallet/cli"
)

func main() {
	c := client.NewCmdClient("http://localhost:8545", "./keystore")
	c.Run()
}
