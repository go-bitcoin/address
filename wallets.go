package address

import (
	"fmt"
	"os"
	"log"
	"io/ioutil"
	"encoding/gob"
	"crypto/elliptic"
	"bytes"
)

type Wallets struct {
	Wallets map[string]*Wallet
}

//新建钱包夹，用于保存多个钱包
func NewWallets() (*Wallets, error) {
	wallets := make(map[string]*Wallet)
	walletFile := Wallets{wallets}
	err := walletFile.LoadFromFile()
	return &walletFile, err
}

//新建一个钱包
func (ws *Wallets) CreateWallet() string {
	wallet := NewWallet()
	address := fmt.Sprintf("%s", wallet.GetAddress())
	ws.Wallets[address] = wallet
	return address
}

//从钱包夹中获取所有钱包地址
func (ws *Wallets) GetAddresses() []string {
	var addresses []string
	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}
	return addresses
}

//根据钱包地址，获取钱包
func (ws *Wallets) GetWallet(address string) Wallet {
	return *ws.Wallets[address]
}
func (ws *Wallets) LoadFromFile() error {
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}
	fileContent, err := ioutil.ReadFile(walletFile)
	if err != nil {
		log.Panic(err)
	}
	var wallets Wallets
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	decoder.Decode(&wallets)
	ws.Wallets = wallets.Wallets
	return nil
}
func (ws Wallets) SaveToFile() {
	var content bytes.Buffer
	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws)
	if err != nil {
		log.Panic(err)
	}
	err = ioutil.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}
