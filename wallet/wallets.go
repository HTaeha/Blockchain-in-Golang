package wallet

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

const walletFile = "./tmp/wallets.data"

// Wallets : map형태로 badgerDB에 저장.
// key : address
// value : Wallet의 주소값
type Wallets struct {
	Wallets map[string]*Wallet
}

// CreateWallets : Wallets를 생성한다.
// Wallets 껍데기를 만들고 그 안에 기존 Wallets를 로드한다.
func CreateWallets() (*Wallets, error) {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)

	err := wallets.LoadFile()

	return &wallets, err
}

// AddWallet : Wallet을 추가한다.
func (ws *Wallets) AddWallet() string {
	wallet := MakeWallet()
	// byte를 string으로 변환.
	address := fmt.Sprintf("%s", wallet.Address())

	ws.Wallets[address] = wallet

	return address
}

// GetAllAddresses : Wallet 안의 모든 주소들을 반환.
func (ws *Wallets) GetAllAddresses() []string {
	var addresses []string

	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}

	return addresses
}

// GetWallet : map 이어서 접근하기 쉬움.
func (ws Wallets) GetWallet(address string) Wallet {
	return *ws.Wallets[address]
}

// LoadFile : 저장된 Wallets를 불러온다.
func (ws *Wallets) LoadFile() error {
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}

	var wallets Wallets

	fileContent, err := ioutil.ReadFile(walletFile)
	if err != nil {
		return err
	}

	// elliptic.P256 알고리즘으로 디코딩
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)
	if err != nil {
		return err
	}

	ws.Wallets = wallets.Wallets

	return nil
}

// SaveFile : elliptic.P256 알고리즘을 이용해서 walletFile에 저장한다.
func (ws *Wallets) SaveFile() {
	var content bytes.Buffer

	// elliptic.P256 알고리즘으로 인코딩
	gob.Register(elliptic.P256())

	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws)
	if err != nil {
		log.Panic(err)
	}

	// 0644 : permission, 파일이 이미 존재하지 않으면 생성한다.
	err = ioutil.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}
