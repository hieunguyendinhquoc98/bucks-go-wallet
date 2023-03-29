package models

import (
	"bytes"
	"encoding/gob"
	"github.com/bucks-go-wallet/utils"
)

type TxInput struct {
	ID        []byte //refer transaction that output inside it
	Out       int    // how many outputs?
	Signature []byte // used for the output pubkey
	PubKey    []byte
}

type TxOutput struct {
	Value      int    //token
	PubKeyHash []byte //needed to unlock token inside Value field
}

type TxOutputs struct {
	Outputs []TxOutput
}

func NewTxOutput(value int, address string) *TxOutput {
	txo := &TxOutput{value, nil}
	txo.Lock([]byte(address))
	return txo
}

func (in *TxInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := PublicKeyHash(in.PubKey)

	return bytes.Compare(lockingHash, pubKeyHash) == 0
}

func (out *TxOutput) Lock(address []byte) {
	pubKeyHash := utils.Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4] //remove checksum, version byte
	out.PubKeyHash = pubKeyHash
}

func (out *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

func (outs *TxOutputs) Serialize() []byte {
	var buffer bytes.Buffer
	encode := gob.NewEncoder(&buffer)
	err := encode.Encode(outs)
	utils.Handle(err)
	return buffer.Bytes()
}

func DeserializeOutputs(outputs []byte) TxOutputs {
	var txOutputs TxOutputs
	decode := gob.NewDecoder(bytes.NewReader(outputs))
	err := decode.Decode(&txOutputs)
	utils.Handle(err)
	return txOutputs
}
