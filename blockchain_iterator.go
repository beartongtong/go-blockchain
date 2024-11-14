package main

import (
	"log"

	"github.com/boltdb/bolt"
)

// BlockchainIterator is used to iterate over blockchain blocks
type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

// Next 这段代码是 `BlockchainIterator` 结构体的 `Next` 方法，用于获取区块链上的下一个区块。
//下面是这个方法的功能和步骤解释：
//1. 创建一个变量 `block`，用于存储从数据库中读取的区块。
//2. 使用数据库事务进行查看操作，打开区块链数据库的 `blocksBucket` 存储桶，并根据当前哈希获取对应的序列化区块数据。
//3. 使用 `DeserializeBlock` 函数对序列化的区块数据进行反序列化，得到区块对象 `block`。
//4. 更新迭代器的 `currentHash` 为当前区块的前一个区块哈希，以便在下一次迭代时获取前一个区块。
//5. 返回获取的区块对象 `block`。
//总的来说，这个方法的目的是在区块链上迭代获取下一个区块，它会从数据库中查找当前哈希对应的区块数据，
//将其反序列化为区块对象，并将迭代器的当前哈希更新为下一个区块的前一个区块哈希。这样，每次调用 `Next` 方法就可以获取区块链上的下一个区块。
func (i *BlockchainIterator) Next() *Block {
	var block *Block

	// 在此添加检查，确保 i.db 是有效的数据库连接
	if i.db == nil {
		log.Panic("Database connection is nil")
	}
	if i.currentHash == nil {
		// 如果 i.currentHash 为 nil，可能需要返回 nil 或者进行其他处理
		return nil
	}

	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		block = DeserializeBlock(encodedBlock)

		return nil
	})

	if err != nil {
		log.Panic(err)
	}
	if block == nil {
		// Block is nil, skip to the next iteration
		return i.Next()
	}
	i.currentHash = block.PrevBlockHash

	return block
}
