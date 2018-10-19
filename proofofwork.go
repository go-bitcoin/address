package address

import (
	"bytes"
	"crypto/sha256"
	"math"
	"math/big"
	"fmt"
)

const targetBit = 12

type ProofOfWork struct {
	block  *Block
	target *big.Int
}

func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBit))
	return &ProofOfWork{b, target}
}
func (pow *ProofOfWork) PrepareData(nonce int) []byte {
	var result []byte
	b := pow.block
	result = bytes.Join([][]byte{
		b.PreviousBlockHash,
		IntToHex(b.Timestamp),
		b.HashTransactions(),
		IntToHex(int64(targetBit)),
		IntToHex(int64(nonce))},
		[]byte{})
	return result
}
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var blockHash [32]byte
	nonce := 1
	for nonce < math.MaxInt64 {
		data := pow.PrepareData(nonce)
		blockHash = sha256.Sum256(data)
		fmt.Printf("mining new block...hash:%x\n",blockHash)
		hashInt.SetBytes(blockHash[:])
		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}
	}
	return nonce, blockHash[:]
}
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int
	hashInt.SetBytes(pow.block.Hash)
	if hashInt.Cmp(pow.target) != -1 {
		return false
	}
	data := pow.PrepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	fmt.Printf("%x\n",hash)
	fmt.Printf("%x\n",pow.block.Hash)
	if !bytes.Equal(hash[:], pow.block.Hash) {
		return false
	}
	return true
}
