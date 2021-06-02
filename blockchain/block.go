package blockchain

// BlockChain structure
type BlockChain struct {
	Blocks []*Block
}

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

// AddBlock : data의 값을 가지는 블록을 추가한다.
func (chain *BlockChain) AddBlock(data string) {
	prevBlock := chain.Blocks[len(chain.Blocks)-1]
	new := CreateBlock(data, prevBlock.Hash)
	chain.Blocks = append(chain.Blocks, new)
}

// Genesis : 체인의 맨 처음 블록이다. prevHash 값이 비어있다.
func Genesis() *Block {
	return CreateBlock("Genesis", []byte{})
}

// InitBlockChain : Genesis 블록을 시작으로 하는 블록체인을 생성한다.
func InitBlockChain() *BlockChain {
	return &BlockChain{[]*Block{Genesis()}}
}
