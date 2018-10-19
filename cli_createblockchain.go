package address

import (
	"log"
	"fmt"
)

func (cli *CLI) CreateBlockChain(address string) {
	if !ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := CreateBlockChain(address)
	bc.db.Close()
	fmt.Println("New BlockChain has been Created. Done!")
}
