package hdwallet

import (
	"crypto/ecdsa"
	"fmt"
	hdkeystore "go_wallet/hdkeystore"
	"go_wallet/mnemonic"
	"os"
	"path/filepath"

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

// NewKeyFromMnemonic 从助记词生成ECDSA私钥。
// 参数:
//
//	mn - 助记词字符串。
//
// 返回值:
//
//	*ecdsa.PrivateKey - 如果成功生成私钥，则返回私钥实例，否则返回nil。
//	error - 如果生成私钥过程中出现错误，则返回错误信息。
func NewKeyFromMnemonic(mn string) (*ecdsa.PrivateKey, error) {
	// 使用BIP39生成种子。
	seed, _ := bip39.NewSeedWithErrorChecking(mn, "")
	// 使用种子生成主密钥。
	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, err
	}
	// 解析默认的派生路径。
	path, err := accounts.ParseDerivationPath(defaultDerivationPath)
	if err != nil {
		return nil, err
	}
	// 沿着派生路径生成子密钥。
	for _, n := range path {
		masterKey, err = masterKey.Child(n)
		if err != nil {
			return nil, err
		}
	}
	// 获取ECDSA私钥。
	privateKey, err := masterKey.ECPrivKey()
	privateKeyECDSA := privateKey.ToECDSA()
	if err != nil {
		return nil, err
	}
	return privateKeyECDSA, nil
}

// DerivePublicKey 通过给定的ECDSA私钥派生出对应的公钥。
// 参数:
//
//	privateKey - 指向ECDSA私钥的指针。
//
// 返回值:
//
//	*ecdsa.PublicKey - 如果成功派生公钥，则返回公钥实例，否则返回nil。
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

// StoreKey 将HD钱包的密钥存储到指定的文件中，并使用给定的密码进行加密。
// 参数:
//
//	pass - 用于加密密钥的密码字符串。
//
// 返回值:
//
//	error - 如果存储过程中出现错误，则返回错误信息。
func (wallet HDWallet) StoreKey(pass string) error {
	// 生成密钥文件的完整路径。
	filename := wallet.HDKeyStore.JoinPath(wallet.Address.Hex())
	// 将密钥存储到文件中。
	return wallet.HDKeyStore.StoreKey(filename, &wallet.HDKeyStore.Key, pass)
}

// LoadWallet 从指定的文件中加载HD钱包。
// 参数:
//
//	filename - 密钥文件名。
//	datadir - 存储密钥文件的目录路径。
//
// 返回值:
//
//	HDWallet - 如果成功加载钱包，则返回钱包实例，否则返回空实例。
//	error - 如果加载过程中出现错误，则返回错误信息。
func LoadWallet(filename, datadir string) (HDWallet, error) {
	// 创建一个新的HD密钥库实例。
	hdks := hdkeystore.NewHDkeyStoreNoKey(datadir)
	fmt.Println("Please input password for:", filename)

	// 生成密钥文件的完整路径。
	fullPath := filepath.Join(datadir, filename)
	// 检查文件是否存在。
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		fmt.Printf("File does not exist: %s\n", fullPath)
		return HDWallet{}, fmt.Errorf("file does not exist: %s", fullPath)
	}

	// 从用户输入获取密码。
	pass, _ := gopass.GetPasswd()
	// 将文件名转换为以太坊地址。
	fromaddr := common.HexToAddress(filename)
	// 从密钥文件中获取私钥。
	privateKey, err := hdks.GetKey(fromaddr, fullPath, string(pass)) // 确保使用完整路径
	if err != nil {
		fmt.Println("Failed to get key from keystore:", err)
		return HDWallet{}, err
	}
	// 检查私钥是否为空。
	if privateKey == nil {
		fmt.Println("Private key is nil for address:", fromaddr.Hex())
		return HDWallet{}, fmt.Errorf("private key is nil for address: %s", fromaddr.Hex())
	}
	// 返回加载的HD钱包实例。
	return HDWallet{
		Address:    fromaddr,
		HDKeyStore: hdks,
	}, nil
}

func LoadWalletByPass(filename, datadir, pass string) (HDWallet, error) {
	// 创建一个新的HD密钥库实例。
	hdks := hdkeystore.NewHDkeyStoreNoKey(datadir)
	fmt.Println("Please input password for:", filename)

	// 生成密钥文件的完整路径。
	fullPath := filepath.Join(datadir, filename)
	// 检查文件是否存在。
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		fmt.Printf("File does not exist: %s\n", fullPath)
		return HDWallet{}, fmt.Errorf("file does not exist: %s", fullPath)
	}

	// 将文件名转换为以太坊地址。
	fromaddr := common.HexToAddress(filename)
	// 从密钥文件中获取私钥。
	privateKey, err := hdks.GetKey(fromaddr, fullPath, string(pass)) // 确保使用完整路径
	if err != nil {
		fmt.Println("Failed to get key from keystore:", err)
		return HDWallet{}, err
	}
	// 检查私钥是否为空。
	if privateKey == nil {
		fmt.Println("Private key is nil for address:", fromaddr.Hex())
		return HDWallet{}, fmt.Errorf("private key is nil for address: %s", fromaddr.Hex())
	}
	// 返回加载的HD钱包实例。
	return HDWallet{
		Address:    fromaddr,
		HDKeyStore: hdks,
	}, nil
}
