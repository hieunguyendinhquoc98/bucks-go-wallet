package main

import (
	"fmt"
	"github.com/bucks-go-wallet/models"
	"strconv"
)

func main() {

	chain := models.InitBlockChain()

	chain.AddBlock("First Block after Cody")
	chain.AddBlock("Second Block after Cody")
	chain.AddBlock("Third Block after Cody")

	for _, block := range chain.Blocks {
		fmt.Printf("Previour Hash: %x\n", block.PrevHash)
		fmt.Printf("Data in Block: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)

		pow := models.NewProof(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()
	}
}
