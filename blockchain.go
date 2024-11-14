package main

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
)

const dbFile = "blockchain_%s.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

// Blockchain implements interactions with a DB
type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

// CreateBlockchain 这段代码是一个 `CreateBlockchain` 函数，用于创建一个新的区块链，并返回一个指向该区块链的指针。
//1. 函数接受两个参数：`address`（创世区块的奖励接收地址）和 `nodeID`（节点的标识）。
//2. 首先，根据传入的 `nodeID` 构建数据库文件路径 `dbFile`。
//3. 使用 `dbExists` 函数检查数据库文件是否已经存在，如果存在则输出信息并终止程序。
//4. 创建一个 Coinbase 交易 `cbtx`，其中将创世区块的奖励发送到指定的地址，并使用 `genesisCoinbaseData` 作为交易数据。
//5. 使用 `NewGenesisBlock` 函数创建一个创世区块，并将上面创建的 Coinbase 交易作为其唯一的交易。
//6. 打开一个 BoltDB 数据库文件，如果出现错误则触发 panic。
//7. 在数据库的一个更新事务中，创建一个 bucket（类似于一个命名空间）用于存储区块，并将创世区块的哈希和序列化后的区块数据存储在该 bucket 中。还将创世区块的哈希存储在键为 "l" 的 entry 中，作为链的尖端。
//8. 最后，返回一个指向新创建的区块链的指针。
//总之，这个函数用于创建一个新的区块链，包括一个创世区块，并将创世区块的奖励发送到指定的地址。它还会在数据库中存储创世区块以及与之相关的信息。
func CreateBlockchain(address, nodeID string) *Blockchain {
	dbFile := fmt.Sprintf(dbFile, nodeID)
	if dbExists(dbFile) {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	var tip []byte

	cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
	genesis := NewGenesisBlock(cbtx)

	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Panic(err)
		}

		err = b.Put(genesis.Hash, genesis.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), genesis.Hash)
		if err != nil {
			log.Panic(err)
		}
		tip = genesis.Hash

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}

	return &bc
}

// NewBlockchain 这段代码是用于创建新的区块链实例的函数 `NewBlockchain`。
//下面是这个函数的功能和步骤解释：
//1. 使用给定的 `nodeID` 创建数据库文件名 `dbFile`。
//2. 检查数据库文件是否存在，如果不存在，输出错误信息并退出程序。
//3. 声明一个变量 `tip`，用于存储区块链的顶部块的哈希。
//4. 使用 `bolt.Open` 函数打开数据库文件，返回数据库句柄 `db`。如果出现错误，触发 Panic。
//5. 使用数据库事务更新，打开名为 `blocksBucket` 的 Bucket，然后获取最后一个区块的哈希（保存在 `l` 键下），并将其赋值给 `tip`。
//6. 使用 `defer` 语句确保在函数结束后关闭数据库连接。
//7. 创建一个新的 `Blockchain` 实例，传入顶部块的哈希和数据库句柄。
//8. 返回新创建的区块链实例。
//总的来说，这个函数的目的是根据给定的 `nodeID` 创建一个新的区块链实例。它首先检查数据库文件是否存在，然后打开数据库并获取最后一个区块的哈希作为初始值，然后创建并返回一个新的区块链实例。如果数据库文件不存在，则输出错误信息并退出程序。
func NewBlockchain(nodeID string) *Blockchain {
	//fmt.Println("NewBlockchain")
	//fmt.Println("NewBlockchain-nodeID:", nodeID)
	dbFile := fmt.Sprintf(dbFile, nodeID)
	if dbExists(dbFile) == false {
		fmt.Println("No existing blockchain found. Create one first.")
		//os.Exit(1)
		return nil
	}
	fmt.Println("NewBlockchain-dbFile:", dbFile)
	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err.Error())
	}
	//fmt.Println("NewBlockchain-db:", db)
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))

		return nil
	})
	//fmt.Println("NewBlockchain-tip:", tip)
	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}

	return &bc
}

//NewBlockchain0400 这段代码是用于创建新的区块链实例的函数 `NewBlockchain`。只有读取权限
//下面是这个函数的功能和步骤解释：
func NewBlockchain0400(nodeID string) *Blockchain {
	//fmt.Println("NewBlockchain0400")
	//fmt.Println("NewBlockchain0400-nodeID:", nodeID)
	dbFile := fmt.Sprintf(dbFile, nodeID)
	if dbExists(dbFile) == false {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}
	fmt.Println("NewBlockchain0400-dbFile:", dbFile)
	var tip []byte
	db, err := bolt.Open(dbFile, 0400, nil)
	if err != nil {
		fmt.Println("2")
		log.Panic(err.Error())
	}

	fmt.Println("NewBlockchain-db:", db)
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))

		return nil
	})
	//fmt.Println("NewBlockchain-tip:", tip)
	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}

	return &bc
}

// AddBlock 这段代码是区块链中的一个方法，用于向区块链中添加新的区块。
//1. 首先，通过数据库事务 `bc.db.Update` 开始数据库写入操作。
//2. 在数据库中获取名为 `blocksBucket` 的存储桶（bucket），这是用于存储区块数据的桶。
//3. 检查要添加的区块是否已存在于数据库中（根据区块的哈希值）。如果已存在，则直接返回，不做任何操作。
//4. 如果区块不存在，将要添加的区块序列化为字节数组（通过 `block.Serialize()` 方法）。
//5. 将区块数据存储到数据库中，键为区块的哈希值，值为序列化后的区块数据。
//6. 获取当前链中的最新区块（通过获取名为 "l" 的键来获取最新区块的哈希值）。
//7. 根据区块高度比较新区块和最新区块的高度，如果新区块的高度更高，则更新最新区块的哈希值。
//8. 提交数据库事务，将写入的数据永久保存到数据库中。
//这个方法的目的是确保区块链中的区块是有序的，并且保持最新区块的引用，以便在添加新区块时更新。
func (bc *Blockchain) AddBlock(block *Block) {
	err := bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		blockInDb := b.Get(block.Hash)

		if blockInDb != nil {
			return nil
		}
		fmt.Println("Add New block", hex.EncodeToString(block.Hash))
		//遍历区块中的交易
		//for _, tx := range block.Transactions {
		//	//输出交易
		//	fmt.Println("tx.ID", hex.EncodeToString(tx.ID))
		//	//遍历交易中的输入
		//	for _, vin := range tx.Vin {
		//		//输出交易中的输入
		//		fmt.Println("vin.Txid", hex.EncodeToString(vin.Txid))
		//		fmt.Println("vin.Vout", vin.Vout)
		//	}
		//	//遍历交易中的输出
		//	for _, vout := range tx.Vout {
		//		//输出交易中的输出
		//		fmt.Println("vout.Value", vout.Value)
		//	}
		//}

		blockData := block.Serialize()
		err := b.Put(block.Hash, blockData)
		if err != nil {
			log.Panic(err)
		}

		lastHash := b.Get([]byte("l"))
		lastBlockData := b.Get(lastHash)
		lastBlock := DeserializeBlock(lastBlockData)

		if block.Height > lastBlock.Height {
			err = b.Put([]byte("l"), block.Hash)
			if err != nil {
				log.Panic(err)
			}
			bc.tip = block.Hash
		}
		//UTXOSet := UTXOSet{bc}
		//UTXOSet.Update(block)
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

// FindTransaction finds a transaction by its ID
func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	fmt.Println("-----------------FindTransaction start-----------------------")
	bci := bc.Iterator()
	fmt.Println("-----------------bci := bc.Iterator()-----------------------")
	for {
		block := bci.Next()
		//fmt.Println("-----------------block := bci.Next()-----------------------")
		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	fmt.Println("-----------------FindTransaction end----------------------")
	return Transaction{}, errors.New("Transaction is not found")
}

// FindUTXO finds all unspent transaction outputs and returns transactions with spent outputs removed
func (bc *Blockchain) FindUTXO() map[string]TXOutputs {
	UTXO := make(map[string]TXOutputs)
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				// Was the output spent?
				if spentTXOs[txID] != nil {
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx {
							continue Outputs
						}
					}
				}

				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					inTxID := hex.EncodeToString(in.Txid)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return UTXO
}

// Iterator 这段代码是 `Blockchain` 结构体的方法 `Iterator`，用于创建并返回一个新的区块链迭代器。
//下面是这个方法的功能和步骤解释：
//1. 创建一个新的 `BlockchainIterator` 结构体实例 `bci`，将其初始化为当前区块链的创世块哈希（tip）和关联的数据库。
//2. 返回创建的 `bci` 区块链迭代器实例。
//总的来说，这个方法的目的是为当前区块链创建一个迭代器，以便在区块链上进行迭代遍历操作。迭代器是一种常见的设计模式，用于按顺序访问集合中的元素，这里用于遍历区块链上的每个区块。
func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.db}

	return bci
}

// GetBestHeight returns the height of the latest block
func (bc *Blockchain) GetBestHeight() int {
	var lastBlock Block

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash := b.Get([]byte("l"))
		blockData := b.Get(lastHash)
		lastBlock = *DeserializeBlock(blockData)

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return lastBlock.Height
}

// GetBlock finds a block by its hash and returns it
func (bc *Blockchain) GetBlock(blockHash []byte) (Block, error) {
	var block Block

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		blockData := b.Get(blockHash)

		if blockData == nil {
			return errors.New("Block is not found.")
		}

		block = *DeserializeBlock(blockData)

		return nil
	})
	if err != nil {
		return block, err
	}

	return block, nil
}

// GetBlockHashes 这个 `GetBlockHashes` 方法是 `Blockchain` 结构体的一个方法，用于获取区块链中所有区块的哈希值列表。让我解释一下这个方法的功能：
//1. 创建一个空的切片 `blocks`，用于存储区块的哈希值。
//2. 通过 `bc.Iterator()` 获取一个区块链迭代器 `bci`，该迭代器初始化指向最新的区块。
//3. 进入循环，通过 `bci.Next()` 获取当前迭代器指向的区块。
//4. 将当前区块的哈希值 `block.Hash` 添加到 `blocks` 切片中。
//5. 检查当前区块的 `PrevBlockHash` 是否为空。如果为空，说明已经到达区块链的创世块（Genesis Block），退出循环。
//6. 如果 `PrevBlockHash` 不为空，则继续迭代，将迭代器指向上一个区块，重复步骤 3~5。
//7. 返回存储了所有区块哈希值的 `blocks` 切片。
//这个方法的主要目的是获取整个区块链中每个区块的哈希值，并以切片的形式返回。通常在处理网络同步、区块链浏览等场景中，我们需要获取区块的哈希值列表。
func (bc *Blockchain) GetBlockHashes() [][]byte {
	var blocks [][]byte
	bci := bc.Iterator()

	for {
		block := bci.Next()

		blocks = append(blocks, block.Hash)

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return blocks
}

// MineBlock 这段代码是 `Blockchain` 结构体的方法 `MineBlock`，用于挖掘一个新的区块并将其添加到区块链中。
//下面是这个方法的功能和步骤解释：
//1. 遍历传入的交易列表 `transactions`，对每个交易进行验证。如果交易无效，则触发 Panic。
//2. 通过读取区块链数据库，获取最后一个区块的哈希和高度。
//3. 使用 `NewBlock` 函数创建一个新的区块，传入当前待确认的交易列表 `transactions`、最后一个区块的哈希和高度。
//4. 使用数据库事务更新，将新的区块的序列化数据存储到区块链数据库中，同时更新最后一个区块的哈希。将区块链结构体的 `tip` 指向新挖掘的区块的哈希。
//5. 返回新挖掘的区块对象。
//总的来说，这个方法的目的是根据给定的交易列表，挖掘一个新的区块并将其添加到区块链中。在挖掘过程中，它会验证交易的有效性，并将新区块的数据存储到数据库中。
func (bc *Blockchain) MineBlock(transactions []*Transaction) *Block {
	var lastHash []byte
	var lastHeight int

	for _, tx := range transactions { //1. 遍历传入的交易列表 `transactions`，对每个交易进行验证。如果交易无效，则触发 Panic。
		// TODO: ignore transaction if it's not valid
		if bc.VerifyTransaction(tx) != true {
			log.Panic("ERROR: Invalid transaction")
		}
	}

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		blockData := b.Get(lastHash) //2. 通过读取区块链数据库，获取最后一个区块的哈希和高度。
		block := DeserializeBlock(blockData)

		lastHeight = block.Height //2. 通过读取区块链数据库，获取最后一个区块的哈希和高度。

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(transactions, lastHash, lastHeight+1, 0, []byte("MineBlock"))
	//3. 使用 `NewBlock` 函数 进行POW运算，创建一个新的区块，传入当前待确认的交易列表 `transactions`、最后一个区块的哈希和高度。
	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}

		bc.tip = newBlock.Hash

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return newBlock
}

//commitTransaction 这段代码是 `Blockchain` 结构体的方法 `commitTransaction`，用于将交易添加到区块链中。
//下面是这个方法的功能和步骤解释：
//1. 首先，它会遍历交易列表 `transactions`，对每个交易进行验证。如果交易无效，则触发 Panic。
//2. 然后，它会创建一个新的区块，并将交易列表 `transactions` 传入 `NewBlock` 函数。
//3. 最后，它会将新创建的区块添加到区块链中，通过调用 `AddBlock` 方法。
//总的来说，这个方法的目的是将交易添加到区块链中。它会创建一个新的区块，并将交易列表传入 `NewBlock` 函数，然后将新创建的区块添加到区块链中。
func (bc *Blockchain) commitTransaction(transactions []*Transaction, data string) *Block {
	var lastHash []byte
	var lastHeight int
	fmt.Println("commitTransaction")
	for _, tx := range transactions { //1. 遍历传入的交易列表 `transactions`，对每个交易进行验证。如果交易无效，则触发 Panic。
		// TODO: ignore transaction if it's not valid
		//if bc.VerifyTransaction(tx) != true {
		//	log.Panic("ERROR: Invalid transaction  Txid:", hex.EncodeToString(tx.ID))
		//}
		if tx == nil {
			log.Panic("ERROR: Invalid transaction  Txid:", hex.EncodeToString(tx.ID))
		}
	}
	fmt.Println("err := bc.db.View(func(tx *bolt.Tx) error {")
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		blockData := b.Get(lastHash) //2. 通过读取区块链数据库，获取最后一个区块的哈希和高度。
		block := DeserializeBlock(blockData)

		lastHeight = block.Height //2. 通过读取区块链数据库，获取最后一个区块的哈希和高度。

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	fmt.Println("newBlock := NewBlock(transactions, lastHash, lastHeight+1, 1)")
	newBlock := NewBlock(transactions, lastHash, lastHeight+1, 1, []byte(data))
	//3. 使用 `NewBlock` 函数 进行POW运算，创建一个新的区块，传入当前待确认的交易列表 `transactions`、最后一个区块的哈希和高度。
	fmt.Println("err = bc.db.Update")
	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}

		bc.tip = newBlock.Hash

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	fmt.Println("return newBlock")
	return newBlock
}

// SignTransaction signs inputs of a Transaction
func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

// VerifyTransaction 这段代码看起来是一个区块链（Blockchain）上的一个方法，用于验证交易（Transaction）。以下是代码的主要功能：
//1. 首先，它检查传入的交易 `tx` 是否是一个Coinbase交易（Coinbase交易是区块链中的特殊类型交易，用于奖励矿工）。如果是Coinbase交易，
//它会直接返回`true`，因为Coinbase交易不需要验证。
//2. 如果不是Coinbase交易，它会创建一个空的映射 `prevTXs`，用于存储前一个交易的引用。
//3. 接下来，它遍历 `tx` 中的每一个输入（`vin`），并尝试查找每个输入所引用的前一笔交易（prevTX）。
//这是通过调用 `bc.FindTransaction(vin.Txid)` 来实现的，其中 `vin.Txid` 是交易输入引用的前一笔交易的ID。
//4. 如果找到了前一笔交易 `prevTX`，它将前一笔交易的ID（以十六进制编码的形式）作为键，将前一笔交易本身作为值存储在 `prevTXs` 映射中。
//5. 最后，它调用 `tx.Verify(prevTXs)` 方法，将 `prevTXs` 映射传递给交易 `tx` 的 `Verify` 方法，
//以进行交易验证。`tx.Verify` 方法将使用前一笔交易的信息来验证当前交易的有效性，包括验证输入的数字签名等。
//总的来说，这段代码用于检查一个交易是否在给定的区块链上是有效的，通过查找和验证交易输入引用的前一笔交易来完成这一过程。
//如果交易是Coinbase交易，它将直接被视为有效。否则，它将尝试查找前一笔交易，并将前一笔交易的信息传递给 `tx.Verify` 方法进行验证。
func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	fmt.Println("------------------VerifyTransaction start----------")
	if tx.IsCoinbase() {
		return true
	}

	prevTXs := make(map[string]Transaction)
	fmt.Println("------------------prevTXs := make(map[string]Transaction)----------", tx.Vin)
	for _, vin := range tx.Vin {
		fmt.Println("------------------VerifyTransaction vin.Txid----------", vin.Txid)
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}
	fmt.Println("------------------VerifyTransaction end----------")
	return tx.Verify(prevTXs)
}

//dbExists函数接受一个参数 dbFile，表示要检查的数据库文件的路径。
//使用 os.Stat 函数检查文件是否存在。如果文件不存在，会返回一个 os.IsNotExist 错误，表示文件不存在。
//如果文件不存在，则返回 false，表示数据库文件不存在。
//如果文件存在，则返回 true，表示数据库文件存在。
func dbExists(dbFile string) bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}
