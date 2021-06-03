package blockchain

import (
	"fmt"

	"github.com/dgraph-io/badger/v3"
)

const (
	// database Path.
	dbPath = "./tmp/blocks"
)

// BlockChain structure
// 마지막 해쉬를 저장하고 DB 포인터를 저장해서 블록을 관리.
type BlockChain struct {
	LastHash []byte
	Database *badger.DB
}

// BlockChainIterator : DB에 저장된 블록체인을 순회하기 위해 생성.
type BlockChainIterator struct {
	// 현재 가리키고 있는 hash
	CurrentHash []byte
	Database    *badger.DB
}

// InitBlockChain : Genesis 블록을 시작으로 하는 블록체인을 생성한다.
func InitBlockChain() *BlockChain {
	var lastHash []byte

	opts := badger.DefaultOptions(dbPath)
	// store key and metadata
	opts.Dir = dbPath
	// store all values
	opts.ValueDir = dbPath

	db, err := badger.Open(opts)
	Handle(err)

	// Update : read and write transactions on our databases.
	// Txn : Transaction, which can be read-only or read-write.
	err = db.Update(func(txn *badger.Txn) error {
		// Blockchain이 비어있다는 뜻. (KeyNotFound)
		// lh(last hash) key 가 없음.
		if _, err := txn.Get([]byte("lh")); err == badger.ErrKeyNotFound {
			fmt.Println("No existing blockchain found")
			// 검증된 Genesis Block을 생성.
			genesis := Genesis()
			fmt.Println("Genesis proved")

			// transaction에 저장.
			// genesis.Hash -> genesis.Seriaize()
			err = txn.Set(genesis.Hash, genesis.Serialize())
			Handle(err)
			// lh -> genesis.Hash
			err = txn.Set([]byte("lh"), genesis.Hash)

			lastHash = genesis.Hash

			return err
			// lh key가 존재하는 경우.
		} else {
			// lh 키의 value를 lastHash에 할당.
			item, err := txn.Get([]byte("lh"))
			Handle(err)
			lastHash, err = item.ValueCopy(nil)
			return err
		}
	})

	Handle(err)

	blockchain := BlockChain{lastHash, db}
	return &blockchain
}

// AddBlock : data의 값을 가지는 블록을 추가한다.
func (chain *BlockChain) AddBlock(data string) {
	var lastHash []byte

	// View : read-only type of transaction
	err := chain.Database.View(func(txn *badger.Txn) error {
		// 마지막 해쉬를 찾는다.
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, err = item.ValueCopy(nil)

		return err
	})
	Handle(err)

	// 이전 해쉬값과 데이터로 새로운 블록을 생성.
	newBlock := CreateBlock(data, lastHash)

	err = chain.Database.Update(func(txn *badger.Txn) error {
		// 새로운 블록의 해쉬값을 저장한다.
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		Handle(err)
		// lh 키값을 새로운 블록의 해쉬값으로 바꾼다.
		err = txn.Set([]byte("lh"), newBlock.Hash)

		// BlockChain의 마지막 해쉬를 새로운 블록의 해쉬값으로 지정한다.
		chain.LastHash = newBlock.Hash

		return err
	})
	Handle(err)
}

// Iterator : 블록체인 이터레이터. LastHash부터 이전 해쉬로 가며 순회할 수 있다.
func (chain *BlockChain) Iterator() *BlockChainIterator {
	iter := &BlockChainIterator{chain.LastHash, chain.Database}

	return iter
}

// Next : BlockChain의 다음 블록을 반환한다.
func (iter *BlockChainIterator) Next() *Block {
	var block *Block

	err := iter.Database.View(func(txn *badger.Txn) error {
		// 현재 가리키고 있는 hash를 deserialize해서 Block을 복원한다.
		item, err := txn.Get(iter.CurrentHash)
		Handle(err)
		encodedBlock, err := item.ValueCopy(nil)
		block = Deserialize(encodedBlock)

		return err
	})
	Handle(err)

	// iter가 이전 해시를 가리키도록 한다.
	iter.CurrentHash = block.PrevHash

	return block
}
