package main

import (
	"github.com/bucks-go-wallet/models"
	cli2 "github.com/bucks-go-wallet/models/cli"
	"github.com/dgraph-io/badger"
	"os"
)

func main() {

	defer os.Exit(0)
	chain := models.InitBlockChain()
	defer func(Database *badger.DB) {
		err := Database.Close()
		if err != nil {

		}
	}(chain.Database)

	cli := cli2.CommandLine{BlockChain: chain}
	cli.Run()
}
