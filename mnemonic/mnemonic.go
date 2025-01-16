package mnemonic

import (
	"fmt"
	"log"

	"github.com/tyler-smith/go-bip39"
)

func CreateMnemonic() (string, error) {
	b, err := bip39.NewEntropy(128)
	if err != nil {
		log.Panic(err)
	}
	mn, err := bip39.NewMnemonic(b)
	if err != nil {
		log.Panic(err)
	}
	fmt.Println(mn)
	return mn, err
}
