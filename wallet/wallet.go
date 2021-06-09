package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"

	"golang.org/x/crypto/ripemd160"
)

const (
	// checksum length in byte.
	// 4byte length
	checksumLength = 4
	// 0 byte를 16진법으로 표현.
	version = byte(0x00)
)

// Wallet : PrivateKey와 PublicKey를 가지고 있는 Wallet structure.
type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

// Address : version, publicKeyHash, checksum 3가지를 concatenate한 후에 Base58로 인코딩해서 address를 만든다.
func (w Wallet) Address() []byte {
	pubHash := PublicKeyHash(w.PublicKey)

	versionedHash := append([]byte{version}, pubHash...)
	checksum := Checksum(versionedHash)

	fullHash := append(versionedHash, checksum...)
	address := Base58Encode(fullHash)

	fmt.Printf("pub key: %x\n", w.PublicKey)
	fmt.Printf("pub hash: %x\n", pubHash)
	fmt.Printf("address: %s\n", address)

	fmt.Printf("fullHash: %x\n", fullHash)
	fmt.Printf("checksum: %x\n", checksum)
	fmt.Printf("version: %x\n", versionedHash)

	return address
}

// Address: 1EwsppsVck2B5Ndf7nPyHi8uYtEhK44ndm
// FullHash: 0098fa85f9e6b04827b2ce6db7e8e18ff8e9882c99eefb8610
// [Version] 00
// [Pub Key Hash] 98fa85f9e6b04827b2ce6db7e8e18ff8e9882c99
// [CheckSum] eefb8610
// FullHash = Version + Pub Key Hash + CheckSum

// ValidateAddress : Address의 유효성을 검증한다.
// 1 : address의 checksum 부분 (address를 base58로 디코드하여 나온 배열에서 마지막 checksumLength만큼의 부분)
// 2 : address의 version부분과 pubKeyHash 부분을 뽑아서 만든 checksum 부분.
// 1과 2가 같은지 확인.
func ValidateAddress(address string) bool {
	// 1 : address의 checksum 부분
	pubKeyHash := Base58Decode([]byte(address))
	actualChecksum := pubKeyHash[len(pubKeyHash)-checksumLength:]

	// 2 : address의 version부분과 pubKeyHash 부분을 뽑아서 만든 checksum 부분.
	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-checksumLength]
	targetChecksum := Checksum(append([]byte{version}, pubKeyHash...))

	return bytes.Compare(actualChecksum, targetChecksum) == 0
}

// NewKeyPair : 새로운 키페어를 만든다.
// ecdsa라는 비대칭키 암호 알고리즘을 사용.
func NewKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()

	// private key는 랜덤하게 뽑혀진 256bit의 숫자이다.
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}

	// X, Y가 concatenate 되어 pub이 됨.
	// private key로부터 public key를 생성함.
	pub := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)
	return *private, pub
}

// MakeWallet : NewKeyPair를 이용해서 Wallet을 만든다.
func MakeWallet() *Wallet {
	private, public := NewKeyPair()
	wallet := Wallet{private, public}

	return &wallet
}

// PublicKeyHash : sha256, ripemd160을 이용해 퍼블릭키를 해쉬로 변환.
// address 생성에 쓰임.
func PublicKeyHash(pubKey []byte) []byte {
	pubHash := sha256.Sum256(pubKey)

	hasher := ripemd160.New()
	_, err := hasher.Write(pubHash[:])
	if err != nil {
		log.Panic(err)
	}

	publicRipMD := hasher.Sum(nil)

	return publicRipMD
}

// Checksum : payload를 sha256을 이용해 해쉬로 변환하고 checksumLength만큼의 바이트만 사용한다.
func Checksum(payload []byte) []byte {
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])

	return secondHash[:checksumLength]
}
