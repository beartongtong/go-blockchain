package main

import (
	"fmt"
	"log"
)

//这段代码是 `CLI` 结构体的 `getBalance` 方法，它用于查询指定地址的余额。
//下面是这个方法的功能和步骤解释：
//1. 使用给定的地址检验是否为有效的地址。如果地址无效，触发 Panic。
//2. 创建一个新的区块链实例，并将指定的 `nodeID` 传递给 `NewBlockchain` 函数。
//3. 创建一个 `UTXOSet` 实例，传入上一步创建的区块链实例。
//4. 使用 `defer` 语句确保在方法结束后关闭数据库连接。
//5. 初始化一个 `balance` 变量用于累计余额。
//6. 将给定的地址解码为公钥哈希，并从中获取真实的公钥哈希部分（去掉版本和校验和）。
//7. 使用 `UTXOSet` 的 `FindUTXO` 方法查找与公钥哈希相关的未花费输出（UTXO）。
//8. 遍历找到的未花费输出，将其价值（`Value`）累加到 `balance` 中。
//9. 使用 `fmt.Printf` 打印出指定地址的余额。
//总的来说，这个方法的目的是查询指定地址的余额。它通过在 UTXO 集合中查找与给定地址相关的未花费输出，然后累加这些未花费输出的价值来计算余额，并输出给定地址的余额信息。
func (cli *CLI) getBalance(address, nodeID string) {
	if !ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := NewBlockchain(nodeID)
	UTXOSet := UTXOSet{bc}
	defer bc.db.Close()

	balance := 0
	pubKeyHash := Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUTXO(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}
