package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/big"
)

// Take the data from the block

// create a counter (nonce) which starts at 0

// create a hash of the data plus the counter

// check the hash to see if it meets a set of requirements

// Requirements:
// The First few bytes must contain 0s

// Difficulty : 채굴하기 위한 문제의 난이도.
// 256bit 중에 Difficulty만큼의 0을 찾는다.
// const Difficulty = 12

// ProofOfWork structure
type ProofOfWork struct {
	Block *Block
	// 정답. Target 보다 작은 값을 찾으면 정답이다.
	Target *big.Int
}

// NewProof : ProofOfWork 객체를 만들어 리턴한다.
func NewProof(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	// Left shift : 아래 식의 결과는 2^(256-Difficulty)가 된다.
	// target은 전체 256비트 중에 왼쪽에 Difficulty-1 만큼의 0이 존재한다. (2진수)
	// 아래 값보다 작은 값이면 0이 Difficulty만큼 존재하는 것이기 때문에 정답이다.
	target.Lsh(target, uint(256-b.Difficulty))

	pow := &ProofOfWork{b, target}

	return pow
}

// InitData : PrevHash, Data, nonce, Difficulty 를 합쳐 데이터를 만든다. (concatenate)
func (pow *ProofOfWork) InitData(nonce, Difficulty int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.Block.PrevHash,
			pow.Block.HashTransactions(),
			ToHex(int64(nonce)),
			ToHex(int64(Difficulty)),
		},
		[]byte{},
	)

	return data
}

// Run : 정답을 찾아 nonce, hash 값을 반환한다.
// difficulty의 값이 클수록 난이도가 어려워진다. (run이 오래 걸린다.)
func (pow *ProofOfWork) Run(difficulty int) (int, []byte) {
	var intHash big.Int
	var hash [32]byte

	nonce := 0

	// MaxInt64 == 2^63 - 1
	for nonce < math.MaxInt64 {
		data := pow.InitData(nonce, difficulty)
		hash = sha256.Sum256(data)

		fmt.Printf("\r%x", hash)
		intHash.SetBytes(hash[:])

		// intHash가 pow.Target보다 작은 값이면 정답.
		if intHash.Cmp(pow.Target) == -1 {
			break
		} else {
			nonce++
		}
	}
	fmt.Println()

	return nonce, hash[:]
}

// Validate : PoW 가 유효한지 검증한다. 매우 간단하게 처리 가능하다.
// Run 할 때보다 매우 쉽고 빠르게 처리된다.
func (pow *ProofOfWork) Validate() bool {
	var intHash big.Int

	// Block의 Nonce값을 이용해 hash 값을 재현한다.
	data := pow.InitData(pow.Block.Nonce, pow.Block.Difficulty)

	hash := sha256.Sum256(data)
	intHash.SetBytes(hash[:])

	return intHash.Cmp(pow.Target) == -1
}

// ToHex : int64를 []byte로 변환.
func ToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}
