package address

import (
	"encoding/gob"
	"bytes"
	"log"
	"time"
	"crypto/sha256"
	"fmt"
)

type Block struct {
	Timestamp         int64
	Transactions      []*Transaction
	PreviousBlockHash []byte
	Hash              []byte
	Nonce             int
}

func NewBlock(transactions []*Transaction, previousHash []byte) *Block {
	block := Block{
		time.Now().Unix(),
		transactions,
		previousHash,
		[]byte{}, 1,
	}
	pow := NewProofOfWork(&block)
	nonce, hash := pow.Run()
	block.Nonce = nonce
	block.Hash = hash[:]
	fmt.Printf("New Block Has Been Mined, hash is: %x\n",hash[:])
	return &block
}
func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase,}, []byte{})
}
func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}
func Deserialize(blockdata []byte) *Block {
	var result Block
	decoder := gob.NewDecoder(bytes.NewReader(blockdata))
	err := decoder.Decode(&result)
	if err != nil {
		log.Panic(err)
	}
	return &result
}
func (b *Block) HashTransactions() []byte {
	var data [][]byte
	for _, tx := range b.Transactions {
		data = append(data, tx.ID)
	}
	txHash := sha256.Sum256(bytes.Join(data, []byte{}))
	return txHash[:]
}
