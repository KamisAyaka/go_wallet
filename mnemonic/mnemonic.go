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
	// 生成128位的熵。
	b, err := bip39.NewEntropy(128)
	if err != nil {
		log.Panic(err)
	}
	// 使用生成的熵创建助记词。
	mn, err := bip39.NewMnemonic(b)
	if err != nil {
		log.Panic(err)
	}
	// 打印生成的助记词。
	fmt.Println(mn)
	// 返回生成的助记词和可能的错误。
	return mn, err
}
