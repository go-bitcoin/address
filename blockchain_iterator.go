package address

import (
	"github.com/boltdb/bolt"
	"log"
)

type BlockChainIterator struct {
	CurrentHash []byte
	db          *bolt.DB
}

func (bci *BlockChainIterator) Next() *Block {
	currentHash := bci.CurrentHash
	var block Block
	err := bci.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		block = *Deserialize(b.Get(currentHash))
		bci.CurrentHash = block.PreviousBlockHash
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return &block
}
