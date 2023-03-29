package cli

import (
	"flag"
	"fmt"
	"github.com/bucks-go-wallet/models"
	"github.com/bucks-go-wallet/utils"
	"github.com/dgraph-io/badger"
	"log"
	"os"
	"runtime"
	"strconv"
)

type CommandLine struct {
	BlockChain *models.BlockChain
}

func (cli *CommandLine) PrintUsage() {
	fmt.Println("Usage:")
	fmt.Println(" getbalance -address ADDRESS - get the balance for an address")
	fmt.Println(" createblockchain -address ADDRESS creates a blockchain and sends cody reward to address")
	fmt.Println(" printchain - Prints the blocks in the chain")
	fmt.Println(" send -from FROM -to TO -amount AMOUNT - Send amount of coins")
	fmt.Println(" createwallet - Create a new wallet")
	fmt.Println(" listaddress - Lists the addresses in our wallet file")
	fmt.Println(" reindexutxo - Rebuild the UTXO set")
}

func (cli *CommandLine) ValidateArgs() {
	if len(os.Args) < 2 {
		cli.PrintUsage()
		runtime.Goexit()
	}
}

func (cli *CommandLine) PrintChain() {
	chain := models.ContinueBlockChain("")
	defer func(Database *badger.DB) {
		err := Database.Close()
		if err != nil {

		}
	}(chain.Database)
	iter := chain.Iterator()
	for {
		block := iter.Next()
		fmt.Printf("Previous Hash: %x\n", block.PrevHash)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := models.NewProof(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}
		fmt.Println()

		if len(block.PrevHash) == 0 {
			break
		}
	}
}

func (cli CommandLine) CreateBlockchain(address string) {
	if !models.ValidateAddress(address) {
		log.Panic("Address is invalid")
	}
	chain := models.InitBlockChain(address)
	err := chain.Database.Close()
	if err != nil {
		return
	}

	UTXOSet := models.UTXOSet{BlockChain: chain}
	UTXOSet.Reindex()
	fmt.Println("Finished create block chain")
}

func (cli CommandLine) GetBalance(address string) {
	if !models.ValidateAddress(address) {
		log.Panic("Address is invalid")
	}
	chain := models.ContinueBlockChain(address)
	UTXOSet := models.UTXOSet{BlockChain: chain}
	defer func(Database *badger.DB) {
		err := Database.Close()
		if err != nil {

		}
	}(chain.Database)

	balance := 0
	pubKeyHash := utils.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUnspentTransactions(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of %s: %d\n", address, balance)
}

func (cli *CommandLine) ListAddresses() {
	wallets, _ := models.CreateWallets()
	addresses := wallets.GetAllAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}
}

func (cli *CommandLine) CreateWallet() {
	wallets, _ := models.CreateWallets()
	address := wallets.AddWallet()
	wallets.SaveFile()

	fmt.Printf("New wallet created on the address: %s\n", address)
}
func (cli *CommandLine) ReindexUTXO() {
	chain := models.ContinueBlockChain("")
	defer func(Database *badger.DB) {
		err := Database.Close()
		if err != nil {

		}
	}(chain.Database)

	UTXUSet := models.UTXOSet{BlockChain: chain}
	UTXUSet.Reindex()

	count := UTXUSet.CountTransactions()

	fmt.Printf("Done, there is %d transactions in this UTXO Set\n", count)
}

func (cli CommandLine) Send(from, to string, amount int) {
	if !models.ValidateAddress(from) {
		log.Panic("From Address is invalid")
	}

	if !models.ValidateAddress(from) {
		log.Panic("To Address is invalid")
	}
	chain := models.ContinueBlockChain(from)
	UTXOSet := models.UTXOSet{BlockChain: chain}
	defer func(Database *badger.DB) {
		err := Database.Close()
		if err != nil {

		}
	}(chain.Database)

	tx := models.NewTransaction(from, to, amount, &UTXOSet)
	block := chain.AddBlock([]*models.Transaction{tx})
	UTXOSet.Update(block)
	fmt.Println("Send transaction successfully")
}

func (cli *CommandLine) Run() {
	cli.ValidateArgs()

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("listaddress", flag.ExitOnError)
	reindexCmd := flag.NewFlagSet("reindexutxo", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	switch os.Args[1] {
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		utils.Handle(err)
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		utils.Handle(err)
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		utils.Handle(err)
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		utils.Handle(err)
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		utils.Handle(err)
	case "listaddress":
		err := listAddressesCmd.Parse(os.Args[2:])
		utils.Handle(err)
	case "reindexutxo":
		err := reindexCmd.Parse(os.Args[2:])
		utils.Handle(err)

	default:
		cli.PrintUsage()
		runtime.Goexit()
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			runtime.Goexit()
		}
		cli.GetBalance(*getBalanceAddress)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			runtime.Goexit()
		}
		cli.CreateBlockchain(*createBlockchainAddress)
	}

	if printChainCmd.Parsed() {
		cli.PrintChain()
	}

	if createWalletCmd.Parsed() {
		cli.CreateWallet()
	}

	if listAddressesCmd.Parsed() {
		cli.ListAddresses()
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			runtime.Goexit()
		}

		cli.Send(*sendFrom, *sendTo, *sendAmount)
	}

	if reindexCmd.Parsed() {
		cli.ReindexUTXO()
	}
}
