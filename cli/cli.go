package client

import (
	"context"
	"flag"
	"fmt"
	"go_wallet/hdwallet"
	"go_wallet/sol"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Client struct {
	network string
	dataDir string
}

const TokenContractAddress = "0xD47497a911aD47731055BDC68718D2814d88Ff9B" //token部署合约之后的地址
var chainID = big.NewInt(1234567)

func NewCmdClient(network, dataDir string) *Client {
	return &Client{
		network: network,
		dataDir: dataDir,
	}
}

func (c *Client) Help() {
	fmt.Println("./go_wallet createwallet -pass PASSWORD --for create new wallet")
	fmt.Println("./go_wallet transfer -from FROM_ADDRESS -toaddr TO_ADDRESS -value VALUE --for transfer from acct to toaddr")
	fmt.Println("./go_wallet balance -from FROM --for get balance of acct")
	fmt.Println("./go_wallet sendtoken -from FROM -toaddr TOADDR -value VALUE --for sendtoken")
	fmt.Println("./go_wallet tokenbalance -from FROM --for get token balance of acct")
	fmt.Println("./go_wallet detail -who WHO --for get tokendetail")
}

func (c Client) Run() {
	if len(os.Args) < 2 {
		c.Help()
		os.Exit(1)
	}
	// createwallet
	cw_cmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	cw_cmd_pass := cw_cmd.String("pass", "", "password for wallet")

	// transfer
	transfer_cmd := flag.NewFlagSet("transfer", flag.ExitOnError)
	transfer_cmd_from := transfer_cmd.String("from", "", "FROM ADDRESS")
	transfer_cmd_toaddr := transfer_cmd.String("toaddr", "", "TO ADDRESS")
	transfer_cmd_value := transfer_cmd.Int64("value", 0, "VALUE")

	// balance
	balance_cmd := flag.NewFlagSet("balance", flag.ExitOnError)
	balance_cmd_from := balance_cmd.String("from", "", "FROM")

	// sendtoken
	sendtoken_cmd := flag.NewFlagSet("sendtoken", flag.ExitOnError)
	sendtoken_cmd_from := sendtoken_cmd.String("from", "", "FROM")
	sendtoken_cmd_toaddr := sendtoken_cmd.String("toaddr", "", "TOADDR")
	sendtoken_cmd_value := sendtoken_cmd.Int64("value", 0, "VALUE")

	// tokenbalance
	tokenbalance_cmd := flag.NewFlagSet("tokenbalance", flag.ExitOnError)
	tokenbalance_cmd_from := tokenbalance_cmd.String("from", "", "FROM")

	// detail
	detail_cmd := flag.NewFlagSet("detail", flag.ExitOnError)
	detail_cmd_who := detail_cmd.String("who", "", "WHO")

	switch os.Args[1] {
	case "createwallet":
		err := cw_cmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("Failed to parse command line arguments", err)
			return
		}
	case "transfer":
		err := transfer_cmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("Failed to parse command line arguments", err)
			return
		}
	case "balance":
		err := balance_cmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("Failed to parse balance_cmd", err)
			return
		}
	case "sendtoken":
		err := sendtoken_cmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("Failed to parse sendtoken_cmd", err)
			return
		}
	case "tokenbalance":
		err := tokenbalance_cmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("Failed to parse tokenbalance_cmd", err)
			return
		}
	case "detail":
		err := detail_cmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("Failed to parse detail_cmd", err)
			return
		}
	}

	if cw_cmd.Parsed() {
		fmt.Println("params is", *cw_cmd_pass)
		c.createWallet(*cw_cmd_pass)
	}

	if transfer_cmd.Parsed() {
		fmt.Println("params is", *transfer_cmd_from, *transfer_cmd_toaddr, *transfer_cmd_value)
		c.transfer(*transfer_cmd_from, *transfer_cmd_toaddr, *transfer_cmd_value)
	}

	if balance_cmd.Parsed() {
		fmt.Println("params is", *balance_cmd_from)
		c.balance(*balance_cmd_from)
	}

	if sendtoken_cmd.Parsed() {
		c.sendtoken(*sendtoken_cmd_from, *sendtoken_cmd_toaddr, *sendtoken_cmd_value)
	}

	if tokenbalance_cmd.Parsed() {
		c.tokenbalance(*tokenbalance_cmd_from)
	}

	if detail_cmd.Parsed() {
		c.tokendetail(*detail_cmd_who)
	}
}

func (c *Client) createWallet(pass string) error {
	w := hdwallet.NewHDWallet(c.dataDir)
	return w.StoreKey(pass)
}

func (c *Client) transfer(from, to string, value int64) error {
	w, _ := hdwallet.LoadWallet(from, c.dataDir)
	cli, _ := ethclient.Dial(c.network)
	defer cli.Close()
	nonce, _ := cli.NonceAt(context.Background(), common.HexToAddress(from), nil)

	gaslimit := uint64(210000)
	gasprice := big.NewInt(5000000000)
	amount := big.NewInt(value)
	tx := types.NewTransaction(nonce, common.HexToAddress(to), amount, gaslimit, gasprice, []byte("Salary"))
	signedTx, err := w.HDKeyStore.SignTx(common.HexToAddress(from), tx, chainID)
	if err != nil {
		fmt.Println("Failed to sign tx", err)
	}
	return cli.SendTransaction(context.Background(), signedTx)
}

func (c *Client) balance(from string) (int64, error) {
	cli, err := ethclient.Dial(c.network)
	if err != nil {
		log.Panic("Failed to connect to Ethereum network")
	}
	defer cli.Close()

	addr := common.HexToAddress(from)
	value, err := cli.BalanceAt(context.Background(), addr, nil)
	if err != nil {
		log.Panic("Failed to get balance", err, from)
	}
	fmt.Println("Balance of", from, "is", value)
	return value.Int64(), nil
}

func (c *Client) sendtoken(from, to string, value int64) error {
	cli, _ := ethclient.Dial(c.network)
	defer cli.Close()

	token, _ := sol.NewToken(common.HexToAddress(TokenContractAddress), cli)

	w, _ := hdwallet.LoadWallet(from, c.dataDir)
	auth, _ := w.HDKeyStore.NewTransactOpts(chainID)
	_, err := token.Transfer(auth, common.HexToAddress(to), big.NewInt(value))
	return err
}

func (c *Client) tokenbalance(from string) (int64, error) {
	cli, err := ethclient.Dial(c.network)
	if err != nil {
		log.Panic("Failed to connect to the Ethereum client", err)
	}
	defer cli.Close()

	// 检查合约地址是否有代码
	code, err := cli.CodeAt(context.Background(), common.HexToAddress(TokenContractAddress), nil)
	if err != nil || len(code) == 0 {
		log.Panic("No contract code at given address")
	}

	token, err := sol.NewToken(common.HexToAddress(TokenContractAddress), cli)
	if err != nil {
		log.Panic("Failed to get token contract", err)
	}
	fromaddr := common.HexToAddress(from)
	opts := bind.CallOpts{
		From: fromaddr,
	}
	value, err := token.BalanceOf(&opts, fromaddr)
	if err != nil {
		log.Panic("Failed to get token balance", err)
	}
	fmt.Printf("%s's token balance: %d\n", from, value.Int64())
	return value.Int64(), nil
}

func (c *Client) tokendetail(who string) error {
	cli, err := ethclient.Dial(c.network)
	if err != nil {
		log.Panic("Failed to connect to the Ethereum client", err)
	}
	defer cli.Close()

	query := ethereum.FilterQuery{
		Addresses: []common.Address{},
		Topics:    [][]common.Hash{{}},
	}
	cAddress := common.HexToAddress(TokenContractAddress)
	topicHash := crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))

	logs, err := cli.FilterLogs(context.Background(), query)
	if err != nil {
		log.Panic("Failed to filter logs", err)
	}

	for _, v := range logs {
		if cAddress == v.Address {
			if len(v.Topics) == 3 {
				if v.Topics[0] == topicHash {
					fromF := v.Topics[1].Bytes()[len(v.Topics[1].Bytes())-20:]
					to := v.Topics[2].Bytes()[len(v.Topics[2].Bytes())-20:]
					val := big.NewInt(0)
					val.SetBytes(v.Data)
					if strings.EqualFold(fmt.Sprintf("0x%x", fromF), who) {
						fmt.Printf(" from : 0x%x\n to : 0x%x\n value : %d\n BlockNumber : %d\n", fromF, to, val.Int64(), v.BlockNumber)
						if strings.EqualFold(fmt.Sprintf("0x%x", to), who) {
							fmt.Printf(" from : 0x%x\n to : 0x%x\n value : %d\n BlockNumber : %d\n", fromF, to, val.Int64(), v.BlockNumber)
						}
					}
				}
			}
		}
	}
	return nil
}
