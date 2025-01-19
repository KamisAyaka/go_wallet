package hdkeystore

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"
	"path/filepath"

	"go_wallet/utils"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

type HDKeyStore struct {
	keysDirPath string
	scryptN     int
	scryptP     int
	Key         keystore.Key
}

// NewHDKeyStore 创建一个新的 HDKeyStore 实例，并使用给定的私钥 ECDSA。
func NewHDKeyStore(keysDirPath string, privateKeyECDSA *ecdsa.PrivateKey) *HDKeyStore {
	id := utils.NewRandom()
	uuid := [16]byte{}
	copy(uuid[:], id)
	key := keystore.Key{
		Id:         uuid,
		Address:    crypto.PubkeyToAddress(privateKeyECDSA.PublicKey),
		PrivateKey: privateKeyECDSA,
	}
	return &HDKeyStore{
		keysDirPath: keysDirPath,
		scryptN:     keystore.StandardScryptN,
		scryptP:     keystore.StandardScryptP,
		Key:         key,
	}
}

// NewHDkeyStoreNoKey 创建一个新的 HDKeyStore 实例，但不包含私钥。
func NewHDkeyStoreNoKey(path string) *HDKeyStore {
	return &HDKeyStore{
		keysDirPath: path,
		scryptN:     keystore.StandardScryptN,
		scryptP:     keystore.StandardScryptP,
		Key:         keystore.Key{},
	}
}

// StoreKey 将密钥存储到指定的文件中，并使用给定的密码进行加密。
func (ks *HDKeyStore) StoreKey(filename string, key *keystore.Key, auth string) error {
	keyjson, err := keystore.EncryptKey(key, auth, ks.scryptN, ks.scryptP)
	if err != nil {
		return err
	}
	return utils.WriteKeyFile(filename, keyjson)
}

// JoinPath 将给定的文件名与密钥存储目录路径连接起来，返回完整的文件路径。
func (ks HDKeyStore) JoinPath(filename string) string {
	if filepath.IsAbs(filename) {
		return filename
	}
	return filepath.Join(ks.keysDirPath, filename)
}

// GetKey 从指定的文件中读取并解密密钥，并验证地址是否匹配。
func (ks *HDKeyStore) GetKey(addr common.Address, filename, auth string) (*keystore.Key, error) {
	keyjson, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	key, err := keystore.DecryptKey(keyjson, auth)
	if err != nil {
		return nil, err
	}
	if key.Address != addr {
		return nil, fmt.Errorf("key content mismatch: have account %x, want %x", key.Address, addr)
	}
	ks.Key = *key
	return key, nil
}

// SignTx 使用当前存储的私钥对交易进行签名，并验证签名者的地址是否匹配。
func (ks *HDKeyStore) SignTx(account common.Address, tx *types.Transaction, chainID *big.Int) (*types.Transaction, error) {
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), ks.Key.PrivateKey)
	if err != nil {
		return nil, err
	}
	// 使用 types.Sender 获取发送者地址
	sender, err := types.Sender(types.NewEIP155Signer(chainID), signedTx)
	if err != nil {
		return nil, err
	}

	if sender != account {
		return nil, fmt.Errorf("signer mismatch: have account %x, want %x", sender.Hex(), account.Hex())
	}

	return signedTx, nil
}

// NewTransactOpts 创建一个新的 TransactOpts 实例，用于交易操作。
func (ks *HDKeyStore) NewTransactOpts(chainID *big.Int) (*bind.TransactOpts, error) {
	opts, err := bind.NewKeyedTransactorWithChainID(ks.Key.PrivateKey, chainID)
	if err != nil {
		return nil, err
	}
	return opts, nil
}
