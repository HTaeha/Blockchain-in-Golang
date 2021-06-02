package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
)

// BlockChain structure
type BlockChain struct {
	blocks []*Block
}

// Block structure
type Block struct {
	Hash     []byte
	Data     []byte
	PrevHash []byte
}

// DeriveHash : Data와 PrevHash를 이용해서 Hash를 생성한다.
func (b *Block) DeriveHash() {
	info := bytes.Join([][]byte{b.Data, b.PrevHash}, []byte{})
	hash := sha256.Sum256(info)
	b.Hash = hash[:]
}

// CreateBlock : data와 prevHash를 받아서 새로운 Hash를 생성한 블록을 생성한다.
func CreateBlock(data string, prevHash []byte) *Block {
	block := &Block{[]byte{}, []byte(data), prevHash}
	block.DeriveHash()
	return block
}

// AddBlock : data의 값을 가지는 블록을 추가한다.
func (chain *BlockChain) AddBlock(data string) {
	prevBlock := chain.blocks[len(chain.blocks)-1]
	new := CreateBlock(data, prevBlock.Hash)
	chain.blocks = append(chain.blocks, new)
}

// Genesis : 체인의 맨 처음 블록이다. prevHash 값이 비어있다.
func Genesis() *Block {
	return CreateBlock("Genesis", []byte{})
}

// InitBlockChain : Genesis 블록을 시작으로 하는 블록체인을 생성한다.
func InitBlockChain() *BlockChain {
	return &BlockChain{[]*Block{Genesis()}}
}

func main() {
	chain := InitBlockChain()

	chain.AddBlock("First Block after Genesis")
	chain.AddBlock("Second Block after Genesis")
	chain.AddBlock("Third Block after Genesis")

	for _, block := range chain.blocks {
		fmt.Printf("Previous Hash: %x\n", block.PrevHash)
		fmt.Printf("Data in Block: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Println()
	}
}
