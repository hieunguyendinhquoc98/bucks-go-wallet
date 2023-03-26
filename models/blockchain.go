package models

import (
	"encoding/hex"
	"fmt"
	"github.com/dgraph-io/badger"
	"os"
	"runtime"
)

const (
	dbPath   = "../tmp/blocks"
	dbFile   = "../tmp/blocks/MANIFEST"
	codyData = "First Transaction from Cody"
)

type BlockChain struct {
	LastHash []byte
	Database *badger.DB
}

type BlockChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

// InitBlockChain create a brand-new blockchain
func InitBlockChain(address string) *BlockChain {
	var lastHash []byte

	if DBExists() {
		fmt.Println("Blockchain already exists")
		runtime.Goexit()
	}

	opts := badger.DefaultOptions(dbPath)
	db, err := badger.Open(opts)
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {
		cbtx := CoinbaseTx(address, codyData)
		cody := Cody(cbtx)
		fmt.Println("Cody created")
		err = txn.Set(cody.Hash, cody.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), cody.Hash)
		lastHash = cody.Hash
		return err
	})
	Handle(err)
	return &BlockChain{lastHash, db}
}

func ContinueBlockChain(address string) *BlockChain {
	if DBExists() == false {
		fmt.Println("No existing blockchain found, create one")
		runtime.Goexit()
	}

	var lastHash []byte
	opts := badger.DefaultOptions(dbPath)
	db, err := badger.Open(opts)
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, _ = item.ValueCopy(nil)
		return err
	})
	Handle(err)

	return &BlockChain{lastHash, db}
}

func (chain *BlockChain) AddBlock(transactions []*Transaction) {
	var lastHash []byte

	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, err = item.ValueCopy(nil)
		return err
	})
	Handle(err)
	newBlock := CreateBlock(transactions, lastHash)

	err = chain.Database.Update(func(txn *badger.Txn) error {
		err = txn.Set(newBlock.Hash, newBlock.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), newBlock.Hash)
		chain.LastHash = newBlock.Hash
		return err
	})
	Handle(err)
}

func (chain *BlockChain) Iterator() *BlockChainIterator {
	return &BlockChainIterator{chain.LastHash, chain.Database}
}

func (i *BlockChainIterator) Next() *Block {
	var block *Block
	var encodedBlock []byte
	err := i.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(i.CurrentHash)
		Handle(err)

		encodedBlock, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}
		block = Deserialize(encodedBlock)
		return nil
	})
	Handle(err)

	i.CurrentHash = block.PrevHash
	return block
}

func (chain *BlockChain) FindUnspentTransactions(address string) []Transaction {
	var unspentTxs []Transaction

	spentTXOs := make(map[string][]int)

	iter := chain.Iterator()
	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)
		Outputs: //label of for loop
			for outIdx, out := range tx.Outputs {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				if out.CanBeUnlocked(address) {
					unspentTxs = append(unspentTxs, *tx)
				}
			}
			if tx.IsCoinbase() == false {
				for _, in := range tx.Inputs {
					if in.CanUnlock(address) {
						inTxID := hex.EncodeToString(in.ID)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Out)
					}
				}
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}
	return unspentTxs
}

func (chain *BlockChain) FindUTXO(address string) []TxOutput {
	var UTXOs []TxOutput
	unspentTransactions := chain.FindUnspentTransactions(address)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Outputs {
			if out.CanBeUnlocked(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

func (chain *BlockChain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {

	unspentOuts := make(map[string][]int)
	unspentTxs := chain.FindUnspentTransactions(address)
	accumulated := 0

Work:
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Outputs {
			if out.CanBeUnlocked(address) && accumulated < amount {
				accumulated += out.Value
				unspentOuts[txID] = append(unspentOuts[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOuts
}

func DBExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}
