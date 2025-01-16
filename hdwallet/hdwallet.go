package hdwallet

import (
	"crypto/ecdsa"
	"fmt"
	hdkeystore "go_wallet/keystore"
	"go_wallet/mnemonic"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/howeyc/gopass"
	"github.com/tyler-smith/go-bip39"
)

const defaultDerivationPath = "m/44'/60'/0'/0/1"

type HDWallet struct {
	Address    common.Address
	HDKeyStore *hdkeystore.HDKeyStore
}

// NewHDWallet 创建一个新的HD钱包。
// 参数:
//
//	keysDirPath - 存储钱包密钥的目录路径。
//
// 返回值:
//
//	*HDWallet - 如果成功创建HD钱包，则返回HD钱包的实例，否则返回nil。
func NewHDWallet(keysDirPath string) *HDWallet {
	// 生成助记词。
	mn, err := mnemonic.CreateMnemonic()
	if err != nil {
		fmt.Println("Error creating mnemonic", err)
		return nil
	}

	// 从助记词生成私钥。
	privateKey, err := NewKeyFromMnemonic(mn)
	if err != nil {
		fmt.Println("Error creating private key from mnemonic", err)
		return nil
	}

	// 从私钥推导出公钥。
	publicKey := DerivePublicKey(privateKey)

	// 从公钥生成以太坊地址。
	address := crypto.PubkeyToAddress(*publicKey)

	// 创建HD密钥库并初始化。
	hdks := hdkeystore.NewHDKeyStore(keysDirPath, privateKey)

	// 返回新的HD钱包实例。
	return &HDWallet{
		Address:    address,
		HDKeyStore: hdks,
	}
}

func NewKeyFromMnemonic(mn string) (*ecdsa.PrivateKey, error) {
	// 实现从助记词生成私钥的逻辑
	seed, _ := bip39.NewSeedWithErrorChecking(mn, "")
	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, err
	}
	path, err := accounts.ParseDerivationPath(defaultDerivationPath)
	if err != nil {
		return nil, err
	}
	for _, n := range path {
		masterKey, err = masterKey.Child(n)
		if err != nil {
			return nil, err
		}
	}
	privateKey, err := masterKey.ECPrivKey()
	privateKeyECDSA := privateKey.ToECDSA()
	if err != nil {
		return nil, err
	}
	return privateKeyECDSA, nil
}

// DerivePublicKey 通过给定的ECDSA私钥派生出对应的公钥。
// 这个函数接受一个指向ECDSA私钥的指针作为参数，并返回一个指向ECDSA公钥的指针。
// 如果公钥无法从私钥正确派生，则返回nil。
func DerivePublicKey(privateKey *ecdsa.PrivateKey) *ecdsa.PublicKey {
	// 从私钥中派生公钥。
	publicKey := privateKey.Public()

	// 将派生的公钥断言为ECDSA公钥类型。
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	// 如果断言失败，说明派生的公钥不是ECDSA类型，返回nil。
	if !ok {
		return nil
	}

	// 返回派生并验证过的ECDSA公钥。
	return publicKeyECDSA
}

func (wallet HDWallet) StoreKey(pass string) error {
	filename := wallet.HDKeyStore.JoinPath(wallet.Address.Hex())
	return wallet.HDKeyStore.StoreKey(filename, pass)
}

func LoadWallet(filename, datadir string) (HDWallet, error) {
	hdks := hdkeystore.NewHDkeyStoreNoKey(datadir)
	fmt.Println("Please input password for:", filename)
	pass, _ := gopass.GetPasswd()
	fromaddr := common.HexToAddress(filename)
	_, err := hdks.GetKey(fromaddr, filename, string(pass))
	if err != nil {
		fmt.Println("Failed to get key from keystore")
	}
	return HDWallet{
		Address:    fromaddr,
		HDKeyStore: hdks,
	}, nil
}
