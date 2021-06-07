package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

// Transaction : 블록에 쓰여질 데이터.
type Transaction struct {
	// hash
	ID []byte
	// Array of input
	Inputs []TxInput
	// Array of output
	Outputs []TxOutput
}

// TxOutput : Transaction Output
// Indivisible - 나눌 수 없는, 불가분의
// Output은 쪼갤 수 없다.
// ex) 500원짜리 물건을 사는데 1000원을 내면 1000원을 반으로 쪼개는 것이 아니라 500원(새로운 아웃풋)을 돌려준다.
// Value만큼의 값을 PubKey가 받는다.
type TxOutput struct {
	// locked value in token
	Value int
	// 공개키 : token을 언락하기 위해 필요하다. (value의 안쪽을 보기 위해)
	// Bitcoin에서는 Pubkey가 복잡한 스크립트 언어로 되어 있다.
	// User's account, address
	PubKey string
}

// TxInput : Transaction Input
// ID에 해당하는 Transaction의 아웃풋에서 Out값 위치에 있는 UTXO를 Sig가 보낸다.
type TxInput struct {
	// Input의 ID
	// 사용할 Transaction의 ID.
	ID []byte
	// Output 의 인덱스.
	// 해당 transaction에서 몇 번째 위치한 output과 연결되어 있는지 알려줌.
	Out int
	// Signature : TxOutput의 PubKey와 비슷한 역할.
	// User's account, address
	Sig string
}

// SetID : transaction의 ID를 만들어 넣어준다.
func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte

	encode := gob.NewEncoder(&encoded)
	err := encode.Encode(tx)
	Handle(err)

	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

// CoinbaseTx : 하나의 인풋과 하나의 아웃풋이 있음. 채굴자가 아웃풋을 받는다.
// to : data 받을 사람의 address
func CoinbaseTx(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Coins to %s", to)
	}

	// 참조하는 TxOutput이 없다.
	txin := TxInput{[]byte{}, -1, data}
	// 100 coin을 to에게 보낸다.
	txout := TxOutput{100, to}

	// Transaction Init
	tx := Transaction{nil, []TxInput{txin}, []TxOutput{txout}}

	return &tx
}

// NewTransaction : from account, to account
// amount : 보내고 싶은 코인의 양
// from이 to에게 amount만큼의 코인을 보낸다.
func NewTransaction(from, to string, amount int, chain *BlockChain) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	acc, validOutputs := chain.FindSpendableOutputs(from, amount)

	// 보내고 싶은 만큼의 코인이 없다.
	if acc < amount {
		// 시간과 에러 문자열을 출력한 뒤 패닉을 발생시킴.
		// 패닉 : 프로그램을 종료시킨다. (런타임 에러)
		// recover 함수를 사용하면 panic 후에 복구할 수 있다. (프로그램이 종료되지 않음.)
		log.Panic("Error: not enough funds")
	}

	// 사용할 수 있는 아웃풋의 인덱스들.
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		Handle(err)

		for _, out := range outs {
			input := TxInput{txID, out, from}
			inputs = append(inputs, input)
		}
	}

	// to address로 amount만큼의 코인을 보낸다.
	outputs = append(outputs, TxOutput{amount, to})

	// 모은 UTXO의 양이 보낼 양보다 크면 거스름돈을 받는다.
	if acc > amount {
		outputs = append(outputs, TxOutput{acc - amount, from})
	}

	tx := Transaction{nil, inputs, outputs}
	tx.SetID()

	return &tx
}

// IsCoinbase : Coinbase 인지 판별한다.
func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs[0].Out == -1
}

// CanUnlock : input의 Sig(address)를 알고 있는 사람만 unlock할 수 있다.
func (in *TxInput) CanUnlock(data string) bool {
	return in.Sig == data
}

// CanBeUnlocked : output의 PubKey(address)를 알고 있는 사람만 unlock할 수 있다.
func (out *TxOutput) CanBeUnlocked(data string) bool {
	return out.PubKey == data
}
