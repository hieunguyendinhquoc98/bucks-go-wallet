package models

import (
	"bytes"
	"encoding/hex"
	"github.com/bucks-go-wallet/utils"
	"github.com/dgraph-io/badger"
)

var (
	utxoPrefix   = []byte("utxo-")
	prefixLength = len(utxoPrefix)
)

type UTXOSet struct {
	BlockChain *BlockChain
}

func (set UTXOSet) Reindex() {
	db := set.BlockChain.Database
	set.DeleteByPrefix(utxoPrefix)

	UTXOs := set.BlockChain.FindUTXO()

	err := db.Update(func(txn *badger.Txn) error {
		for txID, outputs := range UTXOs {
			key, err := hex.DecodeString(txID)
			if err != nil {
				return err
			}

			key = append(utxoPrefix, key...)
			err = txn.Set(key, outputs.Serialize())
			utils.Handle(err)
		}
		return nil
	})
	utils.Handle(err)
}

func (set *UTXOSet) Update(block *Block) {
	db := set.BlockChain.Database

	err := db.Update(func(txn *badger.Txn) error {

		for _, tx := range block.Transactions {
			if tx.IsCoinbase() == false {
				for _, in := range tx.Inputs {
					updatedOuts := TxOutputs{}
					inID := append(utxoPrefix, in.ID...)
					item, err := txn.Get(inID)
					utils.Handle(err)
					v, err := item.ValueCopy(nil)
					utils.Handle(err)
					outs := DeserializeOutputs(v)

					for outIdx, out := range outs.Outputs {
						if outIdx != in.Out {
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
						}
					}
					if len(updatedOuts.Outputs) == 0 {
						utils.Handle(txn.Delete(inID))
					} else {
						utils.Handle(txn.Set(inID, updatedOuts.Serialize()))
					}
				}
			}
			newOutputs := TxOutputs{}
			for _, out := range tx.Outputs {
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}
			txID := append(utxoPrefix, tx.ID...)
			utils.Handle(txn.Set(txID, newOutputs.Serialize()))
		}
		return nil
	})
	utils.Handle(err)

}

func (set *UTXOSet) DeleteByPrefix(prefix []byte) {
	deleteKeys := func(keysForDelete [][]byte) error {
		if err := set.BlockChain.Database.Update(func(txn *badger.Txn) error {
			for _, key := range keysForDelete {
				if err := txn.Delete(key); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	}

	collectSize := 100000
	set.BlockChain.Database.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		keysForDelete := make([][]byte, 0, collectSize)
		keysCollected := 0

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			key, err := it.Item().ValueCopy(nil)
			utils.Handle(err)
			keysForDelete = append(keysForDelete, key)
			keysCollected++

			if keysCollected == collectSize {
				err = deleteKeys(keysForDelete)
				utils.Handle(err)
				keysForDelete = make([][]byte, 0, collectSize)
				keysCollected = 0
			}
		}
		if keysCollected > 0 {
			err := deleteKeys(keysForDelete)
			utils.Handle(err)
		}
		return nil
	})
}

func (set UTXOSet) CountTransactions() int {
	db := set.BlockChain.Database
	counter := 0

	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
			counter++
		}
		return nil
	})
	utils.Handle(err)
	return counter
}

func (set UTXOSet) FindUnspentTransactions(pubkeyHash []byte) []TxOutput {
	var UTXOs []TxOutput

	db := set.BlockChain.Database
	err := db.View(func(txn *badger.Txn) error {

		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)

		defer it.Close()
		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {

			item := it.Item()
			v, err := item.ValueCopy(nil)
			utils.Handle(err)
			outs := DeserializeOutputs(v)

			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(pubkeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}
		return nil
	})
	utils.Handle(err)
	return UTXOs
}

func (set *UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOuts := make(map[string][]int)
	accumulated := 0

	db := set.BlockChain.Database
	err := db.View(func(txn *badger.Txn) error {

		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)

		defer it.Close()
		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
			item := it.Item()
			k := item.Key()
			v, err := item.ValueCopy(nil)
			utils.Handle(err)
			k = bytes.TrimPrefix(k, utxoPrefix)
			txID := hex.EncodeToString(k)
			outs := DeserializeOutputs(v)

			for outIdx, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
					accumulated += out.Value
					unspentOuts[txID] = append(unspentOuts[txID], outIdx)
				}
			}
		}
		return nil
	})
	utils.Handle(err)

	return accumulated, unspentOuts
}
