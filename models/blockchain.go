package models

import (
	"fmt"
	"github.com/dgraph-io/badger"
)

const (
	dbPath = "../tmp/blocks"
)

type BlockChain struct {
	LastHash []byte
	Database *badger.DB
}

type BlockChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

func (c *BlockChain) AddBlock(data string) {
	var lastHash []byte
	err := c.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		_, err = item.ValueCopy(lastHash)
		return err
	})
	Handle(err)
	newBlock := CreateBlock(data, lastHash)

	err = c.Database.Update(func(txn *badger.Txn) error {
		err = txn.Set(newBlock.Hash, newBlock.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), newBlock.Hash)
		c.LastHash = newBlock.Hash
		return err
	})
	Handle(err)
}

// InitBlockChain create a brand-new blockchain
func InitBlockChain() *BlockChain {
	var lastHash []byte

	opts := badger.DefaultOptions(dbPath)
	db, err := badger.Open(opts)
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get([]byte("lh")); err == badger.ErrKeyNotFound {
			fmt.Println("No existing blockchain found ")
			cody := Cody()
			fmt.Println("Cody proved")
			err = txn.Set(cody.Hash, cody.Serialize())
			Handle(err)
			err = txn.Set([]byte("lh"), cody.Hash)
			lastHash = cody.Hash
			return err
		} else {
			item, err := txn.Get([]byte("lh"))
			Handle(err)
			_, err = item.ValueCopy(lastHash)
			return err
		}
	})
	Handle(err)

	blockChain := &BlockChain{lastHash, db}
	return blockChain
}

func (c *BlockChain) Iterator() *BlockChainIterator {
	return &BlockChainIterator{c.LastHash, c.Database}
}

func (i *BlockChainIterator) Next() *Block {
	var block *Block
	var encodedBlock []byte
	err := i.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(i.CurrentHash)
		Handle(err)

		_, err = item.ValueCopy(encodedBlock)
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