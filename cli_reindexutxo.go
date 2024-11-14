package main

import "fmt"

//这段代码是 `CLI` 结构体中的 `reindexUTXO` 方法，用于重新构建 UTXO（未花费输出）集合。
//下面是这个方法的功能和步骤解释：
//1. 根据给定的 `nodeID` 创建一个新的区块链实例 `bc`。
//2. 使用区块链实例 `bc` 创建一个 UTXO 集合实例 `UTXOSet`。
//3. 调用 `UTXOSet.Reindex()` 方法，重新构建 UTXO 集合。这将删除现有的 UTXO 数据并重新构建。
//4. 调用 `UTXOSet.CountTransactions()` 方法，计算并返回 UTXO 集合中的交易数量。
//5. 使用 `fmt.Printf` 打印出重新构建完成的信息，显示 UTXO 集合中的交易数量。
//总的来说，这个方法的目的是重新构建 UTXO 集合，以确保区块链系统的正确性和一致性。在区块链中，UTXO 集合是用于管理未花费的交易输出，它需要在区块链数据发生变化时进行更新和维护。
func (cli *CLI) reindexUTXO(nodeID string) {
	bc := NewBlockchain(nodeID)
	UTXOSet := UTXOSet{bc}
	UTXOSet.Reindex()

	count := UTXOSet.CountTransactions()
	fmt.Printf("Done! There are %d transactions in the UTXO set.\n", count)
}
