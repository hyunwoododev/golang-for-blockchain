package blockchain

import (
	"crypto/sha256"
	"fmt"
	"sync"
)

type Block struct {
	Data     string
	Hash     string
	PrevHash string
}

type blockchain struct {
	block []*Block
}

var b *blockchain
var once sync.Once

func initBlockchain() {
	if b == nil {
		b = &blockchain{}
		genesisBlock := createBlock("Genesis Block")
		b.block = append(b.block, genesisBlock)
	}
} 

func (b *Block) calculateHash() {
	hash := sha256.Sum256([]byte(b.Data + b.PrevHash))
	b.Hash = fmt.Sprintf("%x", hash)
}

func getLastHash() string {
	if b == nil || len(b.block) == 0 {
		return ""
	}
	return b.block[len(b.block)-1].Hash
}

func createBlock(Data string) *Block {
	newBlock := Block{Data, "", getLastHash()}
	newBlock.calculateHash()
	return &newBlock
}

func (b *blockchain) AddBlock(Data string) {
	b.block = append(b.block, createBlock(Data))
}

func GetBlockchain() *blockchain {
	once.Do(initBlockchain)
	return b
}

func (b *blockchain) AllBlocks() []*Block {
	return b.block
}
