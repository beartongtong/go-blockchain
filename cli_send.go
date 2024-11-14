package main

import (
	"fmt"
	"log"
)

//这段代码是一个区块链 CLI（命令行接口）中的 `send` 函数。它用于实现将货币从一个钱包地址发送到另一个钱包地址的操作，并可以选择是否立即挖矿产生新的区块。
//代码的功能和流程：
//1. 首先，`send` 函数接受输入参数，包括发送方地址 `from`、接收方地址 `to`、发送的金额 `amount`、节点 ID `nodeID` 以及一个布尔值 `mineNow`，用于指示是否立即挖矿。
//2. 函数首先使用 `ValidateAddress` 函数验证输入的发送方地址和接收方地址是否有效。如果任一地址无效，函数将使用 `log.Panic` 抛出错误信息并终止程序。
//3. 接下来，函数创建了一个新的区块链 `bc` 和一个 UTXO 集合 `UTXOSet`，然后从钱包文件中加载钱包集合。
//4. 通过给定的发送方地址，从钱包集合中获取发送方的钱包 `wallet`。
//5. 创建一个新的未花费交易输出（UTXO）交易 `tx`，将货币从发送方钱包转移到接收方地址。
//6. 如果 `mineNow` 为 `true`，则表示立即挖矿，将创建一个 coinbase 交易和刚刚创建的交易作为交易列表，并通过挖矿产生一个新的区块。新区块产生后，会调用 `UTXOSet.Update` 更新 UTXO 集合。
//7. 如果 `mineNow` 为 `false`，则表示不立即挖矿，而是将交易发送到已知节点中进行广播。
//8. 最后，无论是立即挖矿还是广播交易，函数都会打印出成功的信息。
//总之，这个 `send` 函数用于在区块链上执行交易操作，可以选择是立即挖矿产生新区块还是广播交易至其他节点。
func (cli *CLI) send(from, to string, amount int, nodeID string, mineNow bool) {
	if !ValidateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc := NewBlockchain(nodeID)
	UTXOSet := UTXOSet{bc}
	defer bc.db.Close()

	wallets, err := NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)

	tx := NewUTXOTransaction(&wallet, to, amount, &UTXOSet)

	if mineNow {
		cbTx := NewCoinbaseTX(from, "")
		txs := []*Transaction{cbTx, tx}

		newBlock := bc.MineBlock(txs)
		UTXOSet.Update(newBlock)
	} else {
		sendTx(knownNodes[0], tx)
	}

	fmt.Println("Success!")
}
