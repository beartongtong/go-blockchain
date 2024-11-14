package main

import (
	"fmt"
	"log"
)

//这段代码是一个 `createBlockchain` 函数，用于创建一个新的区块链，并向指定的地址发送创世区块的奖励。
//1. `createBlockchain` 函数接受两个参数：`address`（要发送奖励的地址）和 `nodeID`（节点的标识）。
//2. 首先，函数使用 `ValidateAddress` 函数验证输入的地址是否有效，如果无效则触发 panic。
//3. 然后，函数调用 `CreateBlockchain` 函数创建一个新的区块链，并传入参数 `address` 和 `nodeID`。创建的区块链会包含一个创世区块，奖励会发送到指定的地址。
//4. 使用 `defer` 关键字，确保在函数结束时关闭区块链数据库连接。
//5. 创建一个 `UTXOSet` 实例，用于管理未花费的交易输出（UTXO）集合。
//6. 调用 `UTXOSet` 的 `Reindex` 方法，该方法会重新建立 UTXO 集合的索引，以便在以后进行交易时能够快速检索未花费的输出。
//7. 最后，输出 "Done!" 表示区块链创建过程完成。
//总之，这个函数用于创建一个新的区块链，并为创世区块奖励发送到指定的地址。在创建过程中，还会重新建立未花费的交易输出集合的索引，以便后续的交易处理。
func (cli *CLI) createBlockchain(address, nodeID string) {
	if !ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := CreateBlockchain(address, nodeID)
	defer bc.db.Close()

	UTXOSet := UTXOSet{bc}
	UTXOSet.Reindex()

	fmt.Println("Done!")
}
