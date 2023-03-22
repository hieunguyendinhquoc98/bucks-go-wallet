package models

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/big"
)

// Take the data from the block

// Create a counter (nonce) which starts at 0

// Create a hash of the data plus the counter

// Check the hash to see if it meets a set of requirements

// Requirements:
// - The first few bytes must contain 0s

const Difficulty = 12

type ProofOfWork struct {
	Block  *Block
	Target *big.Int
}

func NewProof(b *Block) *ProofOfWork {
	target := big.NewInt(1)

	// Left shift...
	target.Lsh(target, uint(256-Difficulty))

	return &ProofOfWork{b, target}
}

func (pow *ProofOfWork) InitData(nonce int) []byte {
	return bytes.Join([][]byte{
		pow.Block.PrevHash,
		pow.Block.Data, ToHex(int64(nonce)),
		ToHex(int64(Difficulty))},
		[]byte{})
}

// Run uses algorithm to get hash and nonce
func (pow *ProofOfWork) Run() (int, []byte) {
	var intHash big.Int
	var hash [32]byte

	nonce := 0

	for nonce < math.MaxInt64 {
		data := pow.InitData(nonce)
		hash = sha256.Sum256(data)

		fmt.Printf("\r%x\n", hash)
		intHash.SetBytes(hash[:])

		if intHash.Cmp(pow.Target) == -1 {
			break
		} else {
			nonce++
		}
	}
	fmt.Println()
	return nonce, hash[:]
}

// Validate checks if pow's hash is valid
// If don't check, we need to regen hash in each pow, take time
func (pow *ProofOfWork) Validate() bool {
	var intHash big.Int
	data := pow.InitData(pow.Block.Nonce)

	hash := sha256.Sum256(data)
	intHash.SetBytes(hash[:])

	return intHash.Cmp(pow.Target) == -1
}

func ToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}
