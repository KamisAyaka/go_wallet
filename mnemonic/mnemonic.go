package mnemonic

import (
	"fmt"
	"log"

	"github.com/tyler-smith/go-bip39"
)

// CreateMnemonic 创建一个新的助记词。
// 参数:
// 返回值:
// 	string - 生成的助记词字符串。
// 	error - 如果生成助记词过程中出现错误，则返回错误信息。
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
