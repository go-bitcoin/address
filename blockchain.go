package address

import (
	"github.com/boltdb/bolt"
	"crypto/ecdsa"
	"os"
	"fmt"
	"log"
	"encoding/hex"
	"bytes"
	"errors"
)

const dbFile = "blockchain.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

type BlockChain struct {
	tip []byte
	db  *bolt.DB
}

func CreateBlockChain(address string) *BlockChain {
	if IsDbExist() {
		fmt.Println("BlockChain Already Exists!")
		os.Exit(1)
	}
	var tip []byte
	cbtx := NewCoinBaseTX(address, genesisCoinbaseData)
	genesis := NewGenesisBlock(cbtx)
	db, err := bolt.Open(dbFile, 0644, nil)
	if err != nil {
		log.Panic(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Panic(err)
		}
		err = b.Put(genesis.Hash, genesis.Serialize())
		if err != nil {
			log.Panic(err)
		}
		err = b.Put([]byte("l"), genesis.Hash)
		if err != nil {
			log.Panic(err)
		}
		return err
	})
	tip = genesis.Hash
	return &BlockChain{tip, db}
}
func NewBlockChain() *BlockChain {
	if !IsDbExist() {
		fmt.Println("BlockChain Not Exist, Please Create a BlockChain first!")
		os.Exit(1)
	}
	var tip []byte
	db, err := bolt.Open(dbFile, 0644, nil)
	if err != nil {
		log.Panic(err)
	}
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return &BlockChain{tip, db}
}
func (bc *BlockChain) FindSpendableOutputs(pubKeyMesh []byte, amount int) (int, map[string][]int) {
	spendableOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(pubKeyMesh)
	acc := 0
Work:
	for _, tx := range unspentTXs {
		txId := hex.EncodeToString(tx.ID)
		for idx, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyMesh) {
				spendableOutputs[txId] = append(spendableOutputs[txId], idx)
				acc = acc + out.Value
				if acc >= amount {
					break Work
				}
			}
		}
	}
	return acc, spendableOutputs
}
func (bc *BlockChain) FindTransaction(ID []byte) (Transaction, error) {
	bci := bc.Iterator()
	for {
		block := bci.Next()
		for _, tx := range block.Transactions {
			txId := tx.ID
			if bytes.Compare(ID, txId) == 0 {
				return *tx, nil
			}
		}
		if len(block.PreviousBlockHash) == 0 {
			break
		}
	}
	return Transaction{}, errors.New("Transaction not found")
}
func (bc *BlockChain) FindUnspentTransactions(pubKeyMesh []byte) []Transaction {
	var unspentTXs []Transaction
	spentTXoutpus := make(map[string][]int)
	bci := bc.Iterator()
	for {
		block := bci.Next()
		for _, tx := range block.Transactions {
			txid := hex.EncodeToString(tx.ID)
		Outputs:
			for outIdx, out := range tx.Vout {
				if spentTXoutpus[txid] != nil {
					for _, spentOutIdx := range spentTXoutpus[txid] {
						if outIdx == spentOutIdx {
							continue Outputs
						}
					}
				}
				if out.IsLockedWithKey(pubKeyMesh) {
					unspentTXs = append(unspentTXs, *tx)
				}
				if !tx.IsCoinBase() {
					for _, in := range tx.Vin {
						if in.UseKey(pubKeyMesh) {
							inTXid := hex.EncodeToString(in.Txid)
							spentTXoutpus[inTXid] = append(spentTXoutpus[inTXid], in.Vout)
						}
					}
				}
			}
		}
		if len(block.PreviousBlockHash) == 0 {
			break
		}
	}
	return unspentTXs
}
func (bc *BlockChain) FindUTXOs(pubKeyMesh []byte) []TXOutput {
	unspentTXs := bc.FindUnspentTransactions(pubKeyMesh)
	var outputs []TXOutput
	for _, tx := range unspentTXs {
		for _, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyMesh) {
				outputs = append(outputs, out)
			}
		}
	}
	return outputs
}
func (bc *BlockChain) Iterator() *BlockChainIterator {
	return &BlockChainIterator{bc.tip, bc.db}
}
func (bc *BlockChain) MineBlock(transactions []*Transaction) {
	var lastHash []byte
	for _, tx := range transactions {
		if !bc.VerifyTransaction(tx) {
			log.Panic("ERROR: invalid transaction!")
		}
	}
	db := bc.db
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	newBlock := NewBlock(transactions, lastHash)
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err = b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}
		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}
		return nil
	})
	bc.tip = lastHash
}
func (bc *BlockChain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)
	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}
	tx.Sign(privKey, prevTXs)
}
func (bc *BlockChain) VerifyTransaction(tx *Transaction) bool {
	prevTXs := make(map[string]Transaction)
	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}
	return tx.Verify(prevTXs)
}
func IsDbExist() bool {
	//如果文件不存在，返回false
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}
