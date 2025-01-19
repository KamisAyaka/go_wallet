# Go Wallet

Go Wallet 是一个用于管理以太坊钱包和代币交易的命令行工具。它支持创建钱包、转账、查询余额、发送代币等功能。代码在本地 geth（版本为 1.13.14）生成的区块链中运行。可以自行修改以连接测试网和主网的区块链。

## 目录

- [安装](#安装)
- [使用](#使用)
  - [创建钱包](#创建钱包)
  - [转账](#转账)
  - [查询余额](#查询余额)
  - [发送代币](#发送代币)
  - [查询代币余额](#查询代币余额)
  - [查询代币交易详情](#查询代币交易详情)
- [API 文档](#api-文档)
- [贡献](#贡献)

## 安装

确保你已经安装了 Go 语言环境。然后克隆项目并安装依赖：

```bash
git clone https://github.com/KamisAyaka/go_wallet.git
cd go-wallet
go mod download
```

## 使用

### 创建钱包

```bash
./go_wallet createwallet -pass PASSWORD
```

### 转账

```bash
./go_wallet transfer -from FROM_ADDRESS -toaddr TO_ADDRESS -value VALUE
```

### 查询余额

```bash
./go_wallet balance -from FROM_ADDRESS
```

### 发送代币

```bash
./go_wallet sendtoken -from FROM_ADDRESS -toaddr TO_ADDRESS -value VALUE
```

### 查询代币余额

```bash
./go_wallet tokenbalance -from FROM_ADDRESS
```

### 查询代币交易详情

```bash
./go_wallet detail -who WHO_ADDRESS
```

## API 文档

### Token 合约

`token.go` 文件中定义了与以太坊 ERC-20 代币合约交互的 API。

- **Allowance**: 查询指定地址的代币授权额度。
- **BalanceOf**: 查询指定地址的代币余额。
- **Symbol**: 查询代币的符号。
- **TotalSupply**: 查询代币的总供应量。
- **Approve**: 授权指定地址使用一定数量的代币。
- **Mint**: 铸造新的代币（仅限合约管理员）。
- **Transfer**: 转移代币到指定地址。
- **TransferFrom**: 从一个地址转移代币到另一个地址（需获得授权）。

### HD 钱包

`hdwallet.go` 文件中定义了与 HD 钱包相关的操作。

- **NewHDWallet**: 创建一个新的 HD 钱包。
- **NewKeyFromMnemonic**: 从助记词生成 ECDSA 私钥。
- **DerivePublicKey**: 从私钥派生公钥。
- **StoreKey**: 将密钥存储到文件中。
- **LoadWallet**: 从文件中加载钱包。

### HD 密钥库

`hdkeystore.go` 文件中定义了与 HD 密钥库相关的操作。

- **NewHDKeyStore**: 创建一个新的 HDKeyStore 实例，并使用给定的私钥 ECDSA。
- **NewHDkeyStoreNoKey**: 创建一个新的 HDKeyStore 实例，但不包含私钥。
- **StoreKey**: 将密钥存储到指定的文件中，并使用给定的密码进行加密。
- **JoinPath**: 将给定的文件名与密钥存储目录路径连接起来，返回完整的文件路径。
- **GetKey**: 从指定的文件中读取并解密密钥，并验证地址是否匹配。
- **SignTx**: 使用当前存储的私钥对交易进行签名，并验证签名者的地址是否匹配。
- **NewTransactOpts**: 创建一个新的 TransactOpts 实例，用于交易操作。

### 客户端

`cli.go` 文件中定义了命令行客户端的接口。

- **Help**: 显示帮助信息。
- **Run**: 执行命令行操作。
- **createWallet**: 创建钱包。
- **transfer**: 转账。
- **balance**: 查询余额。
- **sendtoken**: 发送代币。
- **tokenbalance**: 查询代币余额。
- **tokendetail**: 查询代币详情。

## 贡献

欢迎贡献代码！请遵循以下步骤：

1. Fork 项目。
2. 创建你的特性分支 (`git checkout -b feature/AmazingFeature`)。
3. 提交你的更改 (`git commit -m 'Add some AmazingFeature'`)。
4. 推送到分支 (`git push origin feature/AmazingFeature`)。
5. 打开一个 Pull Request。
