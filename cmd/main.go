package main

import (
	cli2 "github.com/bucks-go-wallet/cli"
	"os"
)

func main() {

	defer os.Exit(0)
	cli := cli2.CommandLine{}
	cli.Run()
}
