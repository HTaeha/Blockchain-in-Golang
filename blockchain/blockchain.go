package blockchain

import (
	"encoding/hex"
	"fmt"
	"os"
	"runtime"

	"github.com/dgraph-io/badger/v3"
)

const (
	// database Path.
	dbPath = "./tmp/blocks"
	// DB가 존재하는지 체크하기 위한 MANIFEST file.
	dbFile      = "./tmp/blocks/MANIFEST"
	genesisData = "First Transaction from Genesis"
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

// DBexists : DB가 존재하는지 체크.
func DBexists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

// InitBlockChain : Genesis 블록을 시작으로 하는 블록체인을 생성한다.
func InitBlockChain(address string) *BlockChain {
	var lastHash []byte

	// DB가 이미 존재하면 종료.
	if DBexists() {
		fmt.Println("Blockchain already exists")
		runtime.Goexit()
	}

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
		// 첫 블록을 Coinbase로 한다.
		cbtx := CoinbaseTx(address, genesisData)
		genesis := Genesis(cbtx)
		fmt.Println("Genesis created")
		err = txn.Set(genesis.Hash, genesis.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), genesis.Hash)

		lastHash = genesis.Hash

		return err
	})

	Handle(err)

	blockchain := BlockChain{lastHash, db}
	return &blockchain
}

// ContinueBlockChain : Blockchain이 이미 존재하고 있을 경우 그 블록체인을 리턴.
// address를 인자로 받고 있지만 이 함수 어디에서도 쓰이지 않는다.
// 현재는 블록체인이 하나만 생성되도록 되어 있어서 address에 아무 값이나 넣어도 상관이 없다.
// DB에 하나의 블록체인만 존재.
func ContinueBlockChain(address string) *BlockChain {
	// DB가 존재하지 않으면 종료.
	if DBexists() == false {
		fmt.Println("No existing blockchain found, create one!")
		runtime.Goexit()
	}

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
		// 마지막 해쉬를 찾는다.
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, err = item.ValueCopy(nil)

		return err
	})
	Handle(err)

	blockchain := BlockChain{lastHash, db}
	return &blockchain
}

// AddBlock : transactions를 가지고 있는 블록을 추가한다.
func (chain *BlockChain) AddBlock(transactions []*Transaction) {
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
	newBlock := CreateBlock(transactions, lastHash)

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

// FindUnspentTransactions : 사용하지 않은 Transaction들을 찾는다.
// Transaction에 사용하지 않은 output이 한개라도 있으면 추가해서 반환한다.
func (chain *BlockChain) FindUnspentTransactions(address string) []Transaction {
	var unspentTxs []Transaction

	// key : string
	// value : []int
	spentTXOs := make(map[string][]int)

	iter := chain.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Outputs {
				// 사용한 TXO인지 체크.
				// 사용했다면 (spentTXOs에 들어있다면) continue
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				// 사용하지 않았고 unlock할 수 있다면 unspentTxs에 추가.
				if out.CanBeUnlocked(address) {
					unspentTxs = append(unspentTxs, *tx)
					// 강의에서는 break문이 없다.
					// tx를 unspentTxs에 추가했다면 다음 tx로 넘어가야하지 않을까?
					// 같은 tx가 여러개 추가되는 현상이 발생할 것 같음.
					break
				}
			}
			// Coinbase가 아닐 때
			if tx.IsCoinbase() == false {
				// Output을 찾기 위해 Input을 돈다.
				// inTxID를 ID로 가지는 Tx 의 in.Out 번째 output은 사용한 output이다.
				for _, in := range tx.Inputs {
					if in.CanUnlock(address) {
						inTxID := hex.EncodeToString(in.ID)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Out)
					}
				}
			}
		}

		// Genesis block 까지 오면 break.
		if len(block.PrevHash) == 0 {
			break
		}
	}
	return unspentTxs
}

// FindUTXO : 사용하지 않은 모든 output을 리턴한다.
// UTXOs의 Value를 모두 더하면 총잔고가 된다.
func (chain *BlockChain) FindUTXO(address string) []TxOutput {
	var UTXOs []TxOutput
	unspentTransactions := chain.FindUnspentTransactions(address)

	// UTXO중에 unlock할 수 있는(해당 address 소유자) 것만 모아서 리턴.
	for _, tx := range unspentTransactions {
		for _, out := range tx.Outputs {
			if out.CanBeUnlocked(address) {
				UTXOs = append(UTXOs, out)
			}

		}
	}
	return UTXOs
}

// FindSpendableOutputs : coin based transaction이 아닌 보통의 transaction을 생성한다.
// address로 amount만큼의 코인을 보낸다.
func (chain *BlockChain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOuts := make(map[string][]int)
	unspentTxs := chain.FindUnspentTransactions(address)
	accumulated := 0

Work:
	// UTX 를 순회
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)

		// 한 transaction의 output을 순회.
		for outIdx, out := range tx.Outputs {
			// unlock할 수 있어야 하고 보내고 싶은 코인의 수가 UTXO로부터 모든 금액보다 클 때
			// 보낼 금액이 부족할 때 (accumulated를 더 증가시켜야 함.)
			if out.CanBeUnlocked(address) && accumulated < amount {
				accumulated += out.Value
				unspentOuts[txID] = append(unspentOuts[txID], outIdx)

				// 금액이 다 모임.
				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOuts
}
