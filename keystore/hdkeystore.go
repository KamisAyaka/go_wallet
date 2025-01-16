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
	keysDirPath     string
	scryptN         int
	scryptP         int
	PrivateKeyECDSA *ecdsa.PrivateKey
}

func NewKeyFromECDSA(privateKeyECDSA *ecdsa.PrivateKey) *keystore.Key {
	id := utils.NewRandom()
	uuid := [16]byte{}
	copy(uuid[:], id)
	key := &keystore.Key{
		Id:         uuid,
		Address:    crypto.PubkeyToAddress(privateKeyECDSA.PublicKey),
		PrivateKey: privateKeyECDSA,
	}
	return key
}

func NewHDKeyStore(keysDirPath string, privateKeyECDSA *ecdsa.PrivateKey) *HDKeyStore {
	return &HDKeyStore{
		keysDirPath:     keysDirPath,
		scryptN:         keystore.StandardScryptN,
		scryptP:         keystore.StandardScryptP,
		PrivateKeyECDSA: privateKeyECDSA,
	}
}

func NewHDkeyStoreNoKey(path string) *HDKeyStore {
	return &HDKeyStore{
		keysDirPath:     path,
		scryptN:         keystore.StandardScryptN,
		scryptP:         keystore.StandardScryptP,
		PrivateKeyECDSA: nil,
	}
}

func (ks *HDKeyStore) StoreKey(filename, auth string) error {
	key := NewKeyFromECDSA(ks.PrivateKeyECDSA)
	filename = ks.JoinPath(filename)
	keyjson, err := keystore.EncryptKey(key, auth, ks.scryptN, ks.scryptP)
	if err != nil {
		return err
	}
	return utils.WriteKeyFile(filename, keyjson)
}

func (ks HDKeyStore) JoinPath(filename string) string {
	if filepath.IsAbs(filename) {
		return filename
	}
	return filepath.Join(ks.keysDirPath, filename)
}

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
	ks.PrivateKeyECDSA = key.PrivateKey
	return key, nil
}

func (ks *HDKeyStore) SignTx(account common.Address, tx *types.Transaction, chainID *big.Int) (*types.Transaction, error) {
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), ks.PrivateKeyECDSA)
	if err != nil {
		return nil, err
	}
	// 使用 types.Sender 函数获取交易的发送者地址
	signer := types.NewEIP155Signer(chainID)
	sender, err := types.Sender(signer, signedTx)
	if err != nil {
		return nil, err
	}
	if sender != account {
		return nil, fmt.Errorf("not authorized to sign this account")
	}
	return signedTx, nil
}

func (ks *HDKeyStore) NewTransactOpts(chainID *big.Int) (*bind.TransactOpts, error) {
	opts, err := bind.NewKeyedTransactorWithChainID(ks.PrivateKeyECDSA, chainID)
	if err != nil {
		return nil, err
	}
	return opts, nil
}
