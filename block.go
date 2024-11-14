package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"time"
)

// Block represents a block in the blockchain
type Block struct {
	Timestamp     int64
	Transactions  []*Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
	Height        int
	Data          []byte
}

// NewBlock 这段代码是用于创建新区块的函数 `NewBlock`。以下是这个函数的关键部分：
//1. 创建一个新的区块（`block`）并初始化它的各个字段：
//- `Timestamp`：当前时间戳，表示区块的创建时间。
//- `Transactions`：传入的交易列表，表示新区块包含的交易。
//- `PrevBlockHash`：前一个区块的哈希值，表示新区块与前一个区块的链接。
//- `Hash`：当前区块的哈希值，初值为空。
//- `Nonce`：工作量证明中的随机数（Nonce），初值为0。
//- `Height`：区块的高度，表示区块在区块链中的位置。
//- `Consensustype`：共识的类型，0是传统POW，1是hotstuff。
//2. 创建一个新的工作量证明（Proof of Work）对象 `pow`，并传入当前区块。
//3. 调用工作量证明的 `Run` 方法，该方法会执行工作量证明算法，寻找有效的 `Nonce` 和区块哈希。
//4. 更新区块的哈希和随机数（`Nonce`）字段，将它们设置为工作量证明找到的有效值。
//5. 返回创建的新区块，其中包含了正确的哈希和随机数，表示该区块已经符合了工作量证明的规则。
//这个函数的目的是创建一个新的区块，并计算出符合工作量证明的哈希和随机数，以便该区块可以被添加到区块链中。
func NewBlock(transactions []*Transaction, prevBlockHash []byte, height int, Consensustype int, Data []byte) *Block {
	fmt.Println("NewBlock")
	fmt.Println("data", Data)
	block := &Block{time.Now().Unix(), transactions, prevBlockHash, []byte{}, 0, height, Data}
	fmt.Println("block", block.Data)
	if Consensustype == 0 {
		pow := NewProofOfWork(block)
		nonce, hash := pow.Run()

		block.Hash = hash[:]
		block.Nonce = nonce
	} else if Consensustype == 1 {
		//hotstuff
		//TODO

		var hash [32]byte
		//nonce值为一个随机的正整数
		nonce := 0
		data := bytes.Join(
			[][]byte{
				block.PrevBlockHash,
				block.HashTransactions(),
				IntToHex(block.Timestamp),
				IntToHex(int64(targetBits)),
				IntToHex(int64(nonce)),
			},
			[]byte{},
		)
		hash = sha256.Sum256(data)
		block.Hash = hash[:]
		block.Nonce = nonce
	}

	return block
}

// NewGenesisBlock creates and returns genesis Block
func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{}, 0, 0, []byte("NewGenesisBlock"))
}

// HashTransactions returns a hash of the transactions in the block
func (b *Block) HashTransactions() []byte {
	var transactions [][]byte

	for _, tx := range b.Transactions {
		transactions = append(transactions, tx.Serialize())
	}
	mTree := NewMerkleTree(transactions)

	return mTree.RootNode.Data
}

// Serialize 这个函数是 `Block` 结构体的方法，用于将区块对象序列化为字节数组。
//以下是这个方法的功能和步骤解释：
//1. 创建一个空的字节缓冲区 `result` 用于存储序列化后的数据。
//2. 使用 `gob.NewEncoder` 创建一个编码器 `encoder`，用于将数据编码为字节序列。
//3. 将区块对象 `b` 通过编码器进行编码，将编码后的结果存储在字节缓冲区 `result` 中。
//4. 如果在编码过程中出现错误，即无法将区块对象编码为字节序列，将触发 Panic。
//5. 返回字节缓冲区 `result` 的内容作为序列化后的区块数据。
//总的来说，这个方法的目的是将区块对象序列化为字节数组，以便将区块数据存储在文件中、通过网络传输或进行其他需要持久化和传输的操作。
func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

// DeserializeBlock 这个函数是用于反序列化字节数组为区块对象的操作。
//以下是这个函数的功能和步骤解释：
//1. 创建一个新的空区块对象 `block`，用于存储反序列化后的数据。
//2. 使用 `gob.NewDecoder` 创建一个解码器 `decoder`，用于将字节数组解码为区块对象。
//3. 将字节数组 `d` 通过解码器进行解码，将解码后的数据存储在区块对象 `block` 中。
//4. 如果在解码过程中出现错误，即无法将字节数组解码为区块对象，将触发 Panic。
//5. 返回区块对象 `block` 的指针，指向反序列化后的区块数据。
//总的来说，这个函数的目的是将字节数组反序列化为区块对象，以便从文件中读取区块数据、接收网络传输的区块数据或进行其他需要解析和操作区块数据的操作。
func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err == io.EOF {
		fmt.Println("DeserializeBlock reached EOF")
		return nil
	} else if err != nil {
		log.Panic("DeserializeBlock:", err)
	}

	return &block
}
