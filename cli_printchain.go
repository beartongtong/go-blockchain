package main

import (
	"fmt"
	"strconv"
)

//这段代码是 `CLI` 结构体的 `printChain` 方法，用于打印区块链中的所有区块和每个区块中的交易信息。
//下面是这个方法的功能和步骤解释：
//1. 使用给定的 `nodeID` 创建一个新的区块链实例 `bc`，并在方法结束时关闭与之关联的数据库。
//2. 通过区块链实例 `bc` 创建一个区块链迭代器 `bci`，以便遍历区块链上的每个区块。
//3. 使用循环遍历区块链上的每个区块：
//- 打印区块的哈希和标识信息。
//- 打印区块的高度和前一个区块的哈希。
//- 创建一个新的工作量证明实例 `pow`，用于验证该区块的工作量。
//- 打印工作量证明的验证结果。
//- 遍历区块中的每笔交易，打印交易信息。
//- 打印换行符，以分隔区块信息。
//4. 如果区块的前一个区块哈希为空，表示已经遍历到创世块，退出循环。
//总的来说，这个方法的目的是遍历区块链并逐个打印每个区块的信息，包括区块的高度、哈希、前一个区块哈希、工作量证明验证结果以及其中的交易信息。这有助于查看区块链的完整结构和交易历史。
func (cli *CLI) printChain(nodeID string) {
	bc := NewBlockchain(nodeID)
	defer bc.db.Close()

	bci := bc.Iterator()

	for {
		block := bci.Next()

		fmt.Printf("============ Block %x ============\n", block.Hash)
		fmt.Printf("Height: %d\n", block.Height)
		fmt.Printf("Prev. block: %x\n", block.PrevBlockHash)
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))
		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}
		fmt.Printf("\n\n")

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}
