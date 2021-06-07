package wallet

import (
	"log"

	"github.com/mr-tron/base58"
)

// 0 O l I + /
// Base58 은 64에서 헷갈리는 6개의 단어를 빼고 58개로 이루어진 인코딩 기법이다.
// Wallet 주소를 헷갈려서 잘못 입력하면 원하는 곳으로 코인이나 데이터를 이동시키지 못 한다.

// Base58Encode : Base58로 인코딩
func Base58Encode(input []byte) []byte {
	encode := base58.Encode(input)

	return []byte(encode)
}

// Base58Decode : Base58로 디코딩
func Base58Decode(input []byte) []byte {
	decode, err := base58.Decode(string(input[:]))
	if err != nil {
		log.Panic(err)
	}

	return decode
}
