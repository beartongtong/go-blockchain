package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

var (
	maxNonce = math.MaxInt64
)

const targetBits = 16

// ProofOfWork represents a proof-of-work
type ProofOfWork struct {
	block  *Block
	target *big.Int
}

// NewProofOfWork 这段代码是用于创建一个新的工作量证明（Proof of Work）实例的函数。
//下面是这个函数的功能和步骤解释：
//1. 创建一个大整数 `target`，初始值为 1。
//2. 使用 `target.Lsh(target, uint(256-targetBits))` 操作，将 `target` 左移（Shift Left）操作，
//移动的位数为 `256 - targetBits`。这是为了创建一个目标哈希，用于工作量证明的挖矿计算。
//3. 使用参数 `b`（一个区块对象）和目标 `target`，创建一个新的工作量证明实例 `pow`。
//4. 返回创建的工作量证明实例 `pow`。
//总的来说，这个函数的目的是根据给定的区块和目标位数，创建一个新的工作量证明实例，用于挖矿过程中的难度计算和验证。
//在比特币和类似的区块链系统中，工作量证明是用于保证区块链安全性的机制之一。
func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	pow := &ProofOfWork{b, target}

	return pow
}

func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.HashTransactions(),
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)

	return data
}

// Run 这段代码是工作量证明（Proof of Work）的核心算法，用于挖矿寻找有效的 `Nonce` 和区块哈希。以下是这个函数 `Run` 的关键部分：
//1. 初始化一个 `hashInt` 变量，用于存储区块哈希的大整数表示。
//2. 定义一个 `hash` 数组，用于存储计算出的哈希值。
//3. 初始化 `nonce` 为0，表示初始的随机数。
//然后进入挖矿循环，该循环会不断尝试不同的随机数值（`Nonce`）来计算区块哈希，直到找到符合工作量证明规则的哈希或达到最大尝试次数（`maxNonce`）为止：
//4. 调用 `prepareData` 函数，根据当前的 `nonce` 值构建要计算哈希的数据块。
//5. 使用 SHA-256 哈希算法对构建的数据块进行哈希计算，结果存储在 `hash` 数组中。
//6. 将计算得到的哈希转换为大整数并存储在 `hashInt` 中。
//7. 检查计算得到的哈希是否小于目标哈希值（`target`）。如果小于目标哈希值，说明找到了有效的 `Nonce`，跳出循环。
//8. 如果计算得到的哈希不小于目标哈希值，则增加 `nonce` 值，继续尝试。
//挖矿过程中，会不断输出当前的哈希值，以展示挖矿进度。一旦找到有效的 `Nonce`，就会跳出循环，并返回找到的 `Nonce` 值和对应的区块哈希。
//这些值将被设置为新区块的哈希和随机数，表示这个区块已经通过工作量证明。
//这个函数的目的是根据当前区块的数据和目标哈希值，通过不断尝试不同的 `Nonce` 值，找到一个使区块哈希满足工作量证明条件的 `Nonce` 值，并返回这个 `Nonce` 和对应的区块哈希。
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	fmt.Printf("Mining a new block")
	for nonce < maxNonce {
		data := pow.prepareData(nonce)

		hash = sha256.Sum256(data)
		fmt.Printf("\r%x", hash)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}
	}
	fmt.Print("\n\n")

	return nonce, hash[:]
}

// Validate validates block's PoW
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}
