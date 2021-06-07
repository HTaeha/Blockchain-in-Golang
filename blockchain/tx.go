package blockchain

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

// CanUnlock : input의 Sig(address)를 알고 있는 사람만 unlock할 수 있다.
func (in *TxInput) CanUnlock(data string) bool {
	return in.Sig == data
}

// CanBeUnlocked : output의 PubKey(address)를 알고 있는 사람만 unlock할 수 있다.
func (out *TxOutput) CanBeUnlocked(data string) bool {
	return out.PubKey == data
}
