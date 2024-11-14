package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"math/big"
	"strings"

	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

const subsidy = 10

// Transaction represents a Bitcoin transaction
type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

var UsedTxId = make(map[string][]byte)

// IsCoinbase checks whether the transaction is coinbase
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

// Serialize returns a serialized Transaction
func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

// Hash returns the hash of the Transaction
func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
}

// Sign signs each input of a Transaction
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinbase() {
		return
	}

	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()

	for inID, vin := range txCopy.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash

		dataToSign := fmt.Sprintf("%x\n", txCopy)

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, []byte(dataToSign))
		if err != nil {
			log.Panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)

		tx.Vin[inID].Signature = signature
		txCopy.Vin[inID].PubKey = nil
	}
}

// String returns a human-readable representation of a transaction
func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.ID))

	for i, input := range tx.Vin {

		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.Txid))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Vout))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PubKey))
	}

	for i, output := range tx.Vout {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}

// TrimmedCopy creates a trimmed copy of Transaction to be used in signing
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, vin := range tx.Vin {
		inputs = append(inputs, TXInput{vin.Txid, vin.Vout, nil, nil})
	}

	for _, vout := range tx.Vout {
		outputs = append(outputs, TXOutput{vout.Value, vout.PubKeyHash})
	}

	txCopy := Transaction{tx.ID, inputs, outputs}

	return txCopy
}

// Verify 这段代码是一个Go语言方法，用于验证一个交易（Transaction）的有效性。它接受一个包含前一笔交易的映射 `prevTXs` 作为参数，并执行以下步骤：
//1. 首先，它检查交易是否是Coinbase交易，如果是，它将直接返回`true`，因为Coinbase交易无需验证。
//2. 对于非Coinbase交易，它遍历交易的每个输入（`vin`），并检查 `prevTXs` 中是否存在与输入引用的前一笔交易相对应的交易。
//如果找不到相应的前一笔交易，它将引发一个日志记录（Panic）错误，指示前一笔交易不正确。
//3. 接下来，它创建一个交易的拷贝 `txCopy`，使用 `TrimmedCopy` 方法，该拷贝不包含输入的签名和公钥信息。
//4. 它使用椭圆曲线P-256（`elliptic.P256()`）来定义椭圆曲线。
//5. 然后，它对交易的每个输入执行以下操作：
//- 将交易拷贝 `txCopy` 中当前输入的签名（`Signature`）和公钥（`PubKey`）字段设置为前一笔交易的输出（Vout）中对应的公钥哈希（PubKeyHash）。
//- 从签名中提取 `r` 和 `s` 值，并从公钥中提取 `x` 和 `y` 值。
//- 生成要验证的数据，通常是交易拷贝的十六进制表示，存储在 `dataToVerify` 变量中。
//- 创建一个ECDSA公钥对象 `rawPubKey`，然后使用它来验证签名的有效性。如果验证失败，返回`false`，表示交易无效。
//- 最后，将交交易拷贝 `txCopy` 中的公钥字段重置为`nil`，以便进行下一个输入的验证。
//6. 如果所有的输入都成功验证，该方法返回`true`，表示交易是有效的。
//总的来说，这段代码用于验证一个交易的有效性，确保交易的输入引用了正确的前一笔交易，并且交易的签名是有效的。这是区块链中重要的一部分，用于确保交易的一致性和安全性。
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for inID, vin := range tx.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash

		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.PubKey)
		x.SetBytes(vin.PubKey[:(keyLen / 2)])
		y.SetBytes(vin.PubKey[(keyLen / 2):])

		dataToVerify := fmt.Sprintf("%x\n", txCopy)

		rawPubKey := ecdsa.PublicKey{curve, &x, &y}
		if ecdsa.Verify(&rawPubKey, []byte(dataToVerify), &r, &s) == false {
			return false
		}
		txCopy.Vin[inID].PubKey = nil
	}

	return true
}

// NewCoinbaseTX 这段代码是一个函数 `NewCoinbaseTX`，用于创建一个 coinbase 交易，即区块链中的首个交易，用于为矿工奖励提供新的货币。
//下面是这个函数的功能和步骤解释：
//1. 如果提供的 `data` 为空，则生成随机数据作为 coinbase 交易的数据。这是为了确保每个 coinbase 交易都有一个唯一的数据，用于识别不同的挖矿尝试。
//2. 创建一个 coinbase 交易输入（`TXInput`）对象。coinbase 交易没有真实的交易输出作为输入，所以这里的交易输入不关联任何交易 ID 和输出索引。
//3. 创建一个 coinbase 交易输出（`TXOutput`）对象，其中金额为预定的挖矿奖励（`subsidy`）并发送给目标地址 `to`。
//4. 创建一个 coinbase 交易（`Transaction`）对象，设置其输入和输出列表，计算交易的哈希值（`ID`）。
//5. 返回创建的 coinbase 交易对象。
//总的来说，这个函数的目的是创建一个 coinbase 交易，用于为矿工提供挖矿奖励，同时也可以携带一些随机数据以确保每个 coinbase 交易的唯一性。
func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		randData := make([]byte, 20)
		_, err := rand.Read(randData)
		if err != nil {
			log.Panic(err)
		}

		data = fmt.Sprintf("%x", randData)
	}

	txin := TXInput{[]byte{}, -1, nil, []byte(data)}
	txout := NewTXOutput(subsidy, to)
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{*txout}}
	tx.ID = tx.Hash()

	return &tx
}

func NewCoinbaseTX2(to, data string, amount int) *Transaction {
	if data == "" {
		randData := make([]byte, 20)
		_, err := rand.Read(randData)
		if err != nil {
			log.Panic(err)
		}

		data = fmt.Sprintf("%x", randData)
	}

	txin := TXInput{[]byte{}, -1, nil, []byte(data)}
	txout := NewTXOutput(amount, to)
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{*txout}}
	tx.ID = tx.Hash()

	return &tx
}

// NewUTXOTransaction 这段代码是一个函数 `NewUTXOTransaction`，用于创建一个新的未花费输出（UTXO）交易。
//下面是这个函数的功能和步骤解释：
//1. 获取钱包的公钥哈希（通过 `HashPubKey` 函数）以及可用的未花费输出列表（`validOutputs`）。
//2. 检查钱包中是否有足够的资金（`amount`）来发送交易，如果资金不足，触发 Panic。
//3. 构建交易的输入列表（`inputs`），遍历有效的未花费输出，为每个输出创建一个交易输入（`TXInput`）。
//4. 构建交易的输出列表（`outputs`），包括一个输出用于发送给目标地址 `to`，以及可能的找零地址输出。
//5. 创建一个交易（`Transaction`）对象，设置其输入和输出列表，计算交易的哈希值（`ID`）。
//6. 使用钱包的私钥对交易进行签名。
//7. 返回创建的交易对象。
//总的来说，这个函数的目的是创建一个新的未花费输出交易，即一个包含输入和输出的交易，其中输出将资金发送给目标地址，
//并可能返回余额到找零地址。创建交易后，还对其进行签名以确保交易的合法性。
func NewUTXOTransaction(wallet *Wallet, to string, amount int, UTXOSet *UTXOSet) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	fmt.Println("")
	fmt.Println("----------NewUTXOTransaction start-----------------------")
	fmt.Println("")
	pubKeyHash := HashPubKey(wallet.PublicKey)
	acc, validOutputs := UTXOSet.FindSpendableOutputs(pubKeyHash, amount)
	fmt.Println("acc", acc)
	fmt.Println("validOutputs", validOutputs)
	fmt.Println("amount", amount)
	if acc < amount {
		//log.Panic("ERROR: Not enough funds")
		fmt.Println("ERROR: Not enough funds")
		return nil
	} else {
		// Build a list of inputs
		for txid, outs := range validOutputs {
			txID, err := hex.DecodeString(txid)
			if err != nil {
				log.Panic("txID, err := hex.DecodeString(txid)", err)
			}

			for _, out := range outs {
				input := TXInput{txID, out, nil, wallet.PublicKey}
				inputs = append(inputs, input)
				UsedTxId[txid] = txID
			}
		}
		fmt.Println("inputs------------------------------")
		//循环遍历打印inputs
		for _, input := range inputs {
			fmt.Println("input.Txid", hex.EncodeToString(input.Txid))
			fmt.Println("input.Vout", input.Vout)
			fmt.Println("----------------------------")
			//fmt.Println("input.Signature", input.Signature)
			//fmt.Println("input.PubKey", input.PubKey)
		}
		//遍历打印UsedTxId
		//for k, v := range UsedTxId {
		//	fmt.Println("UsedTxId", k, v)
		//}
		// Build a list of outputs
		from := fmt.Sprintf("%s", wallet.GetAddress())
		outputs = append(outputs, *NewTXOutput(amount, to))
		if acc > amount {
			outputs = append(outputs, *NewTXOutput(acc-amount, from)) // a change
		}
		fmt.Println("outputs", outputs)
		tx := Transaction{nil, inputs, outputs}
		tx.ID = tx.Hash()
		fmt.Println("tx.ID", tx.ID)
		UTXOSet.Blockchain.SignTransaction(&tx, wallet.PrivateKey)
		fmt.Println("UTXOSet.Blockchain.SignTransaction(&tx, wallet.PrivateKey)")

		fmt.Println("")
		fmt.Println("----------NewUTXOTransaction end-----------------------")
		fmt.Println("")
		return &tx
	}
}

// DeserializeTransaction deserializes a transaction
func DeserializeTransaction(data []byte) Transaction {
	var transaction Transaction

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&transaction)
	if err != nil {
		log.Panic(err)
	}

	return transaction
}
