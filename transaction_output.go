package main

import (
	"bytes"
	"encoding/gob"
	"log"
)

// TXOutput represents a transaction output
type TXOutput struct {
	Value      int
	PubKeyHash []byte
}

// Lock 这个`Lock`方法用于设置交易输出（`TXOutput`）的锁定条件。在比特币和其他加密货币中，锁定条件通常是接收者的公钥哈希。
//1. `func (out *TXOutput) Lock(address []byte)`：这是一个与`TXOutput`结构关联的方法。它接受一个`address`参数，这个参数通常是接收者的地址，但在这个方法内部将被解码为公钥哈希。
//2. `pubKeyHash := Base58Decode(address)`：在这里，`Base58Decode`函数用于将地址解码为字节数组。Base58编码通常用于表示加密货币地址，这里将其解码以获取原始字节数组。
//3. `pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]`：解码后的字节数组通常包含版本字节和校验和。这一行代码将从字节数组中移除版本字节和校验和，留下了公钥哈希。
//4. `out.PubKeyHash = pubKeyHash`：最后，将公钥哈希赋值给`TXOutput`结构的`PubKeyHash`字段，以将其作为锁定条件。
//总的来说，这个方法将接收者的地址解码为公钥哈希，并将其设置为交易输出的锁定条件，以确保只有持有相应私钥的人才能使用这个输出。这是加密货币交易中非常重要的一步，用于确保交易的安全性和正确性。
func (out *TXOutput) Lock(address []byte) {
	pubKeyHash := Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

// IsLockedWithKey 这段 Go 代码是用于检查一个交易输出 (TXOutput) 是否由指定的公钥哈希 (pubKeyHash) 锁定的函数。
//函数的主要逻辑是比较交易输出的 `PubKeyHash` 字段和给定的公钥哈希是否相等。
//具体而言，函数使用 `bytes.Compare` 函数比较两个字节数组，`out.PubKeyHash`
//是交易输出的公钥哈希字段，而 `pubKeyHash` 是传入的参数，即待检查的公钥哈希。
//如果两者相等，`bytes.Compare` 将返回 0，表示相等。因此，整个函数的返回值是 `true` 当且仅当这两个公钥哈希相等。
//这种检查通常在区块链中用于验证一笔交易输出是否属于指定的地址。如果公钥哈希匹配，
//那么这个交易输出就可以被相应地址的所有者解锁和使用。
func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

// NewTXOutput create a new TXOutput
func NewTXOutput(value int, address string) *TXOutput {
	txo := &TXOutput{value, nil}
	txo.Lock([]byte(address))

	return txo
}

// TXOutputs collects TXOutput
type TXOutputs struct {
	Outputs []TXOutput
}

// Serialize serializes TXOutputs
func (outs TXOutputs) Serialize() []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(outs)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

// DeserializeOutputs 这段代码的作用是将序列化的交易输出数据反序列化为 `TXOutputs` 结构体。
//1. `var outputs TXOutputs`：创建一个 `TXOutputs` 类型的变量 `outputs`，用于存储反序列化后的交易输出。
//2. `if len(data) == 0 { ... }`：检查传入的序列化数据是否为空，如果为空，则记录错误并触发 panic。
//3. `dec := gob.NewDecoder(bytes.NewReader(data))`：创建一个新的 `gob` 解码器，使用传入的序列化数据初始化它。
//4. `err := dec.Decode(&outputs)`：使用解码器将数据解码到 `outputs` 变量中。如果解码失败，记录错误并触发 panic。
//5. `return outputs`：返回反序列化后的交易输出。
//整个函数的目的是将二进制数据反序列化为 `TXOutputs` 结构体，以便在程序中使用。如果序列化数据为空或解码失败，函数将触发 panic。
func DeserializeOutputs(data []byte) TXOutputs {
	var outputs TXOutputs

	if len(data) == 0 {
		log.Panic("DeserializeOutputs: Empty data")
	}

	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&outputs)
	if err != nil {
		log.Panic("err := dec.Decode(&outputs)", err)
	}

	return outputs
}
