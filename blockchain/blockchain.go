package blockchain

import (
	"crypto/sha256"
	"fmt"
	"sync"
)

type block struct {
	data string
	hash string
	prevHash string
}

type blockchain struct {
	block []*block
}

var b *blockchain
var once sync.Once


func (b *block) calculateHash() {
	hash := sha256.Sum256([]byte(b.data + b.prevHash))
	b.hash = fmt.Sprintf("%x", hash)
}

func getLastHash() string {
	totalBlocks := len(GetBlockchain().block)
	if totalBlocks == 0 {
		return ""
	}
	return GetBlockchain().block[totalBlocks-1].hash
}

func createBlock(data string) *block {
	newBlock := block{data, "", getLastHash()}
	newBlock.calculateHash()
	return &newBlock
}

func GetBlockchain() *blockchain {
	once.Do(func(){
		if b == nil {
			b = &blockchain{}
			b.block = append(b.block, createBlock("Genesis Block"))
		}
	})

	return b
}