package address

import "bytes"

type TXInput struct {
	Txid      []byte
	Vout      int
	Signature []byte
	PubKey    []byte
}
//交易加密信息，是否由pubKeyHash
func (in *TXInput) UseKey(pubKeyHash []byte) bool {
	lockingData := HashPubKey(in.PubKey)
	return bytes.Compare(lockingData, pubKeyHash) == 0
}
