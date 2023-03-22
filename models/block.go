package models

type BlockChain struct {
	Blocks []*Block
}

type Block struct {
	Hash     []byte
	Data     []byte
	PrevHash []byte //represents last block hash, allow to link block together
	Nonce    int
}

// CreateBlock creates new block
func CreateBlock(data string, prevHash []byte) *Block {
	block := &Block{[]byte{}, []byte(data), prevHash, 0}
	pow := NewProof(block)
	nonce, hash := pow.Run()
	block.Nonce = nonce
	block.Hash = hash
	return block
}

func (c *BlockChain) AddBlock(data string) {
	prevBlock := c.Blocks[len(c.Blocks)-1]
	c.Blocks = append(c.Blocks, CreateBlock(data, prevBlock.Hash))
}

func Cody() *Block {
	return CreateBlock("Cody", []byte{})
}

// InitBlockChain create a brand-new blockchain
func InitBlockChain() *BlockChain {
	return &BlockChain{[]*Block{Cody()}}
}
