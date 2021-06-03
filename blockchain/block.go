package blockchain

import (
	"bytes"
	"encoding/gob"
	"log"
)

// Block structure
type Block struct {
	Hash     []byte
	Data     []byte
	PrevHash []byte
	Nonce    int
}

// CreateBlock : data와 prevHash를 받아서 새로운 Hash를 생성한 블록을 생성한다.
func CreateBlock(data string, prevHash []byte) *Block {
	block := &Block{[]byte{}, []byte(data), prevHash, 0}

	// PoW 조건에 맞는 블록을 생성한다.
	pow := NewProof(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

// Genesis : 체인의 맨 처음 블록이다. prevHash 값이 비어있다.
func Genesis() *Block {
	return CreateBlock("Genesis", []byte{})
}

// Serialize : BadgerDB 에 값을 넣기 위해 byte배열로 바꿔준다.
func (b *Block) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)

	err := encoder.Encode(b)

	Handle(err)

	return res.Bytes()
}

// Deserialize : data를 decoding 해서 Block객체로 바꿔준다.
func Deserialize(data []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(data))

	err := decoder.Decode(&block)

	Handle(err)

	return &block
}

// Handle : Error handling.
func Handle(err error) {
	if err != nil {
		log.Panic(err)
	}
}
