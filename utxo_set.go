package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/boltdb/bolt"
)

const utxoBucket = "chainstate"

// UTXOSet represents UTXO set
type UTXOSet struct {
	Blockchain *Blockchain
}

// FindSpendableOutputs 这段代码定义了一个 `FindSpendableOutputs` 方法，属于 `UTXOSet` 结构体，用于查找可以用于消费的未花费输出（UTXO）。
//以下是这个方法的功能和步骤解释：
//1. 创建一个用于存储未花费输出的映射 `unspentOutputs`，其中键是交易 ID（txID），值是未花费输出的索引列表。
//2. 创建一个变量 `accumulated` 用于跟踪已累积的总金额。
//3. 获取 `UTXOSet` 所关联的区块链数据库（`u.Blockchain.db`）。
//4. 在数据库事务中进行查找操作。遍历 UTXO Bucket 中的每个键值对。
//5. 对于每个键值对，将键（交易 ID）转换为十六进制字符串形式作为 `txID`。
//6. 反序列化值（输出列表）为 `outs`。
//7. 遍历输出列表，检查每个输出是否被给定的公钥哈希所锁定，并且累积的金额小于指定的 `amount`。
//8. 如果满足条件，将输出的索引添加到对应的交易 ID 列表中，更新累积金额。
//9. 返回累积的金额 `accumulated` 和未花费输出的映射 `unspentOutputs`。
//总的来说，这个方法的目的是查找可以用于消费的未花费输出（UTXO），以及累积的金额是否满足给定的数量。
//在比特币交易中，发送者需要选择足够的未花费输出来支付交易金额，这个方法用于帮助找到合适的未花费输出。
func (u UTXOSet) FindSpendableOutputs(pubkeyHash []byte, amount int) (int, map[string][]int) {
	//fmt.Println("")
	//fmt.Println("----------FindSpendableOutputs start-----------------------")
	//fmt.Println("")
	unspentOutputs := make(map[string][]int)
	accumulated := 0
	//db := u.Blockchain.db

	err := u.Blockchain.db.View(func(tx *bolt.Tx) error {
		for tx.Writable() {
			//暂停10ms
			time.Sleep(time.Millisecond * 10)
			fmt.Println("等待数据库解锁")
		}
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			txID := hex.EncodeToString(k)
			outs := DeserializeOutputs(v)
			//fmt.Println("FindSpendableOutputs-txID", txID)
			//fmt.Println("FindSpendableOutputs-outs", outs)
			for outIdx, out := range outs.Outputs {
				if out.IsLockedWithKey(pubkeyHash) && accumulated < amount && UsedTxId[txID] == nil {
					accumulated += out.Value
					unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
					UsedTxId[txID] = v
					//fmt.Println("FindSpendableOutputs-outIdx", outIdx)
					//fmt.Println("FindSpendableOutputs-out.Value", out.Value)
				}
			}
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	//fmt.Println("")
	//fmt.Println("----------FindSpendableOutputs end-----------------------")
	//fmt.Println("")
	return accumulated, unspentOutputs
}

// FindUTXO 这段代码是 `UTXOSet` 结构体的方法 `FindUTXO`，用于查找给定公钥哈希（pubKeyHash）对应的未花费交易输出（UTXO）。
//下面是这个方法的功能和步骤解释：
//1. 获取 `UTXOSet` 所关联的区块链数据库（`u.Blockchain.db`）。
//2. 使用数据库事务以只读方式（`db.View`）访问 UTXO Bucket。
//3. 获取 UTXO Bucket 中的游标（`Cursor`），开始遍历其中的键值对。
//4. 遍历 UTXO Bucket 中的每个键值对（k，v），其中 k 是交易 ID，v 是序列化后的未花费交易输出集合。
//5. 对每个未花费交易输出集合 `v` 进行反序列化，以获取其中的多个未花费交易输出。
//6. 对于每个未花费交易输出 `out`，检查它是否使用给定的公钥哈希进行加锁（IsLockedWithKey 方法）。
//7. 如果交易输出被给定的公钥哈希锁定，则将该交易输出 `out` 添加到 `UTXOs` 列表中。
//8. 最终返回存储了满足条件的未花费交易输出的 `UTXOs` 列表。
//总的来说，这个方法的目的是在 UTXO 集合中查找并返回与给定公钥哈希对应的未花费交易输出。在区块链中，
//未花费交易输出表示尚未被花费的币值，通过遍历 UTXO 集合中的每个交易输出，检查是否被给定的公钥哈希锁定，以实现余额计算和交易验证。
func (u UTXOSet) FindUTXO(pubKeyHash []byte) []TXOutput {
	var UTXOs []TXOutput
	db := u.Blockchain.db
	//fmt.Println("FindUTXO-db", db)
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		//fmt.Println("FindUTXO-b", b)
		c := b.Cursor()
		//fmt.Println("FindUTXO-c", c)
		for k, v := c.First(); k != nil; k, v = c.Next() {
			//fmt.Println("FindUTXO-k", hex.EncodeToString(k))
			//fmt.Println("FindUTXO-v", v)
			outs := DeserializeOutputs(v)
			//fmt.Println("FindUTXO-PubKeyHash", hex.EncodeToString(outs.Outputs[0].PubKeyHash))
			//fmt.Println("FindUTXO-outs", outs.Outputs[0].Value)
			for _, out := range outs.Outputs {
				//fmt.Println("FindUTXO-out.Value", out.Value)
				//fmt.Println("FindUTXO-outs.Outputs", outs.Outputs)
				if out.IsLockedWithKey(pubKeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	//遍历UTXOs
	//for _, UTXO := range UTXOs {
	//	fmt.Println("UTXO.Value", UTXO.Value)
	//}
	//fmt.Println("UTXOs", UTXOs)
	return UTXOs
}

// CountTransactions 这段代码是 `UTXOSet` 结构体的 `CountTransactions` 方法，用于计算 UTXO（未花费输出）集合中的交易数量。
//下面是这个方法的功能和步骤解释：
//1. 获取与 `UTXOSet` 相关的区块链数据库实例 `db`。
//2. 初始化一个计数器 `counter`，用于统计交易数量。
//3. 使用数据库事务进行查询操作（`db.View`），访问 UTXO Bucket。
//4. 创建一个游标 `c`，遍历 UTXO Bucket 中的每个键（交易 ID）。
//5. 在循环中，每次遍历一个键（交易 ID）时，将计数器 `counter` 增加一。
//6. 返回计数器 `counter`，表示 UTXO 集合中的交易数量。
//总的来说，这个方法的目的是计算 UTXO 集合中的交易数量。在区块链中，交易数量可以用于监控区块链的活动、统计交易情况等用途。
func (u UTXOSet) CountTransactions() int {
	db := u.Blockchain.db
	counter := 0

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			counter++
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return counter
}

// Reindex 这段代码是一个方法 ，属于 `UTXOSet` 结构体的方法，用于重建 UTXO 集合（未花费输出）。
//1. 获取 `UTXOSet` 所关联的区块链数据库（`u.Blockchain.db`）。
//2. 定义用于存储 UTXO 的 Bucket 名称为 `utxoBucket`。
//3. 使用数据库事务更新，首先删除现有的 UTXO Bucket。如果出现错误，且错误不是 `bolt.ErrBucketNotFound`（表示 Bucket 不存在），则触发 Panic。
//4. 创建一个新的 UTXO Bucket。
//5. 获取 UTXO 集合（未花费输出）。
//6. 使用数据库事务更新，将 UTXO 集合中的每个未花费输出（以交易 ID 为键）序列化后存储在 UTXO Bucket 中。
//总的来说，这个 `Reindex` 方法的目的是重建 UTXO 集合。它首先删除现有的 UTXO Bucket，然后创建一个新的 Bucket，并将 UTXO 集合中的未花费输出按照交易 ID 存储在这个 Bucket 中，以便稍后在验证交易和计算余额时使用。
func (u UTXOSet) Reindex() {
	db := u.Blockchain.db            //1. 获取 `UTXOSet` 所关联的区块链数据库（`u.Blockchain.db`）。
	bucketName := []byte(utxoBucket) //2. 定义用于存储 UTXO 的 Bucket 名称为 `utxoBucket`。

	//3.使用数据库事务更新
	err := db.Update(func(tx *bolt.Tx) error {
		//首先删除现有的 UTXO Bucket。如果出现错误，且错误不是 `bolt.ErrBucketNotFound`（表示 Bucket 不存在），则触发 Panic。
		err := tx.DeleteBucket(bucketName)
		if err != nil && err != bolt.ErrBucketNotFound {
			log.Panic(err)
		}
		//4. 创建一个新的 UTXO Bucket。
		_, err = tx.CreateBucket(bucketName)
		if err != nil {
			log.Panic(err)
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	//5. 获取 UTXO 集合（未花费输出）。
	UTXO := u.Blockchain.FindUTXO()
	//6. 使用数据库事务更新，将 UTXO 集合中的每个未花费输出（以交易 ID 为键）序列化后存储在 UTXO Bucket 中。
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)

		for txID, outs := range UTXO {
			key, err := hex.DecodeString(txID)
			if err != nil {
				log.Panic(err)
			}

			err = b.Put(key, outs.Serialize())
			fmt.Println("txID", txID)
			fmt.Println("outs", outs)
			if err != nil {
				log.Panic(err)
			}
		}

		return nil
	})
}

// Update 这段代码是 `UTXOSet` 结构体的方法 `Update`，用于更新 UTXO 集合（未花费输出）以反映新的区块的交易。
//下面是这个方法的功能和步骤解释：
//1. 获取与 UTXO 集合关联的区块链数据库。
//2. 使用数据库事务更新，遍历新区块中的每个交易。
//3. 对于非 Coinbase 交易，遍历交易的输入（vin）列表。对于每个输入，从数据库中获取对应的输出（outs），并更新输出列表（updatedOuts），排除当前输入所引用的输出。
//4. 如果更新后的输出列表为空，则从数据库中删除对应的交易 ID。
//5. 否则，将更新后的输出列表存储回数据库。
//6. 遍历交易的输出（vout）列表，将每个输出存储到数据库中，使用交易 ID 作为键。
//总的来说，这个方法的目的是根据新的区块中的交易信息，更新 UTXO 集合中的数据，以反映区块链的最新状态。
//它会处理输入（花费）和输出（未花费）的关系，删除已经花费的输出，同时添加新的输出。
func (u UTXOSet) Update(block *Block) {
	//fmt.Println("UTXOSet.Update block", block)
	db := u.Blockchain.db
	//fmt.Println("UTXOSet.Update-db", db)
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		//打印b中的数据
		//err := b.ForEach(func(k, v []byte) error {
		//	fmt.Printf("b.ForEach(func(k, v []byte) error -------Key: %s\n", hex.EncodeToString(k))
		//	return nil
		//})
		//
		//if err != nil {
		//	log.Panic(err)
		//}
		for _, tx := range block.Transactions {
			//fmt.Println("UTXOSet.Update-tx", tx)
			if tx.IsCoinbase() == false {
				for _, vin := range tx.Vin {
					updatedOuts := TXOutputs{}
					hexString := hex.EncodeToString(vin.Txid)

					byteData, err := hex.DecodeString(hexString)
					if err != nil {
						fmt.Println("解码失败:", err)
					}
					outsBytes := b.Get(vin.Txid)
					if outsBytes == nil {
						fmt.Println("hex.EncodeToString(vin.Txid)", hex.EncodeToString(vin.Txid))
						fmt.Println("outsBytes为空 : ", outsBytes)
						if UsedTxId[hex.EncodeToString(vin.Txid)] != nil {
							fmt.Println(hex.EncodeToString(vin.Txid), "已经被使用过了")
						} else {
							fmt.Println(hex.EncodeToString(vin.Txid), "不知道去哪里了")
						}

					}
					fmt.Printf("UTXOSet.Update - vin.Txid: %x, outsBytes: %v\n", byteData, outsBytes)
					outs := DeserializeOutputs(outsBytes)
					//fmt.Println("UTXOSet.Update-outs", outs)
					for outIdx, out := range outs.Outputs {
						if outIdx != vin.Vout {
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
						}
					}
					//fmt.Println("UTXOSet.Update-updatedOuts", updatedOuts)
					if len(updatedOuts.Outputs) == 0 {
						err := b.Delete(vin.Txid)
						if err != nil {
							log.Panic(err)
						}
					} else {
						err := b.Put(vin.Txid, updatedOuts.Serialize())
						if err != nil {
							log.Panic(err)
						}
					}

				}
			}

			newOutputs := TXOutputs{}
			for _, out := range tx.Vout {
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}

			err := b.Put(tx.ID, newOutputs.Serialize())
			if err != nil {
				log.Panic(err)
			}
		}
		fmt.Println("----------UTXOSet.Update-success------------------------")
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}
