package cli

import (
	"flag"
	"fmt"
	"github.com/bucks-go-wallet/models"
	"os"
	"runtime"
	"strconv"
)

type CommandLine struct {
	BlockChain *models.BlockChain
}

func (cli *CommandLine) PrintUsage() {
	fmt.Println("Usage:")
	fmt.Println("add -block BLOCK_DATA - add block of data to the chain")
	fmt.Println("print - Prints the blocks in the chain ")
}

func (cli *CommandLine) ValidateArgs() {
	if len(os.Args) < 2 {
		cli.PrintUsage()
		runtime.Goexit()
	}
}

func (cli *CommandLine) AddBlock(data string) {
	cli.BlockChain.AddBlock(data)
	fmt.Println("Block with data added")
}

func (cli *CommandLine) PrintChain() {
	iter := cli.BlockChain.Iterator()
	for {
		block := iter.Next()
		fmt.Printf("Previour Hash: %x\n", block.PrevHash)
		fmt.Printf("Data in Block: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)

		pow := models.NewProof(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PrevHash) == 0 {
			break
		}
	}
}

func (cli *CommandLine) Run() {
	cli.ValidateArgs()

	addBlockCmd := flag.NewFlagSet("add", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("print", flag.ExitOnError)
	addBlockData := addBlockCmd.String("block", "", "Block data")

	switch os.Args[1] {
	case "add":
		err := addBlockCmd.Parse(os.Args[2:])
		models.Handle(err)
	case "print":
		err := printChainCmd.Parse(os.Args[2:])
		models.Handle(err)
	default:
		cli.PrintUsage()
		runtime.Goexit()
	}

	if addBlockCmd.Parsed() {
		if *addBlockData == "" {
			addBlockCmd.Usage()
			runtime.Goexit()
		}
		cli.AddBlock(*addBlockData)
	}
	if printChainCmd.Parsed() {
		cli.PrintChain()
	}

}
