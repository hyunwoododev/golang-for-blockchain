package main

import (
	"crypto/sha256"
	"fmt"
)

type block struct {
	data string
	hash string
	prevHash string
}

type blockchain struct {
	block []block
}

func (b *blockchain) getLastHash() string {
	if len(b.block) > 0 {
		return b.block[len(b.block)-1].hash
	}
	return ""
}

func (b *blockchain) addBlock(data string){
	newBlock := block{data, "", b.getLastHash()}
	hash := sha256.Sum256([]byte(newBlock.data + newBlock.prevHash))
	newBlock.hash = fmt.Sprintf("%x", hash)
	b.block = append(b.block, newBlock)
}

func (b *blockchain) listBlocks(){
	for _, block := range b.block{
		fmt.Println("Data: ", block.data)
		fmt.Println("Hash: ", block.hash)
		fmt.Println("PrevHash: ", block.prevHash)
	}
}

func main(){
	bc := blockchain{}
	bc.addBlock("First Block")
	bc.addBlock("Second Block")
	bc.addBlock("Third Block")
	bc.addBlock("Fourth Block")
	bc.addBlock("Fifth Block")
	bc.listBlocks()

}