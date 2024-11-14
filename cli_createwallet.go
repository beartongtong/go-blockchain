package main

import "fmt"

//这段代码是 `CLI` 结构体的 `createWallet` 方法，它用于创建一个新的钱包并保存到文件中。
//下面是这个方法的功能和步骤解释：
//1. 使用给定的 `nodeID` 创建一个新的钱包集合（`NewWallets` 方法返回钱包集合和错误，但在此代码中错误被忽略了）。
//2. 调用钱包集合的 `CreateWallet` 方法来创建一个新的钱包，返回一个钱包地址。
//3. 调用钱包集合的 `SaveToFile` 方法，将钱包集合保存到文件中，文件名以 `nodeID` 命名。
//4. 使用 `fmt.Printf` 打印出新创建的钱包地址。
//总的来说，这个方法的目的是创建一个新的钱包，将其保存到文件中，并输出新钱包的地址。钱包是用于存储密钥对和地址的数据结构，在区块链中用于管理账户和签署交易。
func (cli *CLI) createWallet(nodeID string) {
	wallets, _ := NewWallets(nodeID)
	address := wallets.CreateWallet()
	wallets.SaveToFile(nodeID)

	fmt.Printf("Your new address: %s\n", address)
}
