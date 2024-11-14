package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
)

const walletFile = "wallet_%s.dat"

var publicKey ecdsa.PublicKey

// Wallets stores a collection of wallets
type Wallets struct {
	Wallets map[string]*Wallet
}

// NewWallets 这个函数是用于创建一个新的钱包集合（`Wallets`）。下面是这个函数的功能解释：
//1. 初始化一个空的钱包集合 `wallets`。
//2. 为钱包集合创建一个空的钱包映射（`wallets.Wallets`），用于存储每个钱包的地址和对应的钱包实例。
//3. 调用 `LoadFromFile` 方法，从文件中加载已存在的钱包数据到这个钱包集合中。`LoadFromFile` 方法接受一个 `nodeID` 参数，用于确定钱包文件的名称。加载已存在的钱包数据可以保证在创建新钱包之前，现有的钱包数据已被恢复。
//4. 返回创建的钱包集合实例以及可能的错误。如果在加载已存在的钱包数据时出现错误，这个错误会被返回。
//总的来说，这个函数的目的是创建一个新的钱包集合（`Wallets`）并恢复已存在的钱包数据（如果有的话），以便后续可以在其中创建新的钱包。这对于管理区块链中的地址和密钥对非常重要，以便可以正确地签署和验证交易。
func NewWallets(nodeID string) (*Wallets, error) {
	//fmt.Println("NewWallets：nodeID=", nodeID)
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)

	err := wallets.LoadFromFile(nodeID)

	return &wallets, err
}

// CreateWallet 这段代码是 `Wallets` 结构体的 `CreateWallet` 方法，用于创建一个新的钱包并将其添加到钱包集合中。
//以下是这个方法的功能和步骤解释：
//1. 使用 `NewWallet` 函数创建一个新的钱包实例。
//2. 获取新钱包的地址（公钥哈希）。
//3. 将新钱包添加到钱包集合 `ws.Wallets` 中，以钱包地址为键，钱包实例为值。
//4. 返回新钱包的地址。
//总的来说，这个方法的目的是创建一个新的钱包，将其添加到钱包集合中，并返回新钱包的地址。这个地址可以用来接收加密货币或用于其他与区块链交互相关的操作。
func (ws *Wallets) CreateWallet() string {
	wallet := NewWallet()
	address := fmt.Sprintf("%s", wallet.GetAddress())
	fmt.Println("CreateWallet：", address)
	ws.Wallets[address] = wallet

	return address
}

// GetAddresses 这段代码是 `Wallets` 结构体的方法 `GetAddresses`，用于获取钱包集合中存储的所有钱包地址。
//下面是这个方法的功能和步骤解释：
//1. 创建一个空的字符串切片 `addresses`，用于存储钱包地址。
//2. 使用 `range` 循环遍历钱包集合中的每个钱包，其中 `address` 是钱包的地址字符串。
//3. 在每次循环迭代中，将钱包地址 `address` 添加到 `addresses` 切片中。
//4. 循环结束后，返回存储了所有钱包地址的 `addresses` 切片。
//总的来说，这个方法的目的是获取钱包集合中存储的所有钱包地址，并将这些地址存储在一个字符串切片中，以便稍后进行列表显示或其他操作。这在区块链应用中通常用于展示所有钱包的地址列表。
func (ws *Wallets) GetAddresses() []string {
	var addresses []string

	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}
	////for address, wallet := range ws.Wallets {
	////	fmt.Printf("GetAddresses：address=%s\n", address)
	////	fmt.Printf("GetAddresses：wallet.PublicKey=%v\n", wallet.PublicKey)
	////	fmt.Printf("GetAddresses：wallet.PrivateKey=%v\n", wallet.PrivateKey)
	////
	////}
	////找到Wallets中的指定地址
	//wallet := ws.Wallets["18YkgMpHA1qnPFEQm6SZU6JKBpKdMthv5E"]
	//fmt.Printf("GetAddresses：wallet.PublicKey=%v\n", wallet.PublicKey)
	//fmt.Printf("GetAddresses：wallet.PrivateKey=%v\n", wallet.PrivateKey)
	//
	//// 准备要签名的消息
	//message := "Hello, ECDSA!"
	//
	//// 对消息进行哈希
	//hash := sha256.Sum256([]byte(message))
	////ecdsa.PrivateKey 转成 *PrivateKey
	//
	//// 使用私钥进行签名
	//r, s, err := ecdsa.Sign(rand.Reader, &wallet.PrivateKey, hash[:])
	//if err != nil {
	//	fmt.Println("Failed to sign message:", err)
	//	return nil
	//}
	//// 签名成功
	//fmt.Printf("Signature (r, s): (%s, %s)\n", r, s)

	return addresses
}

//获取钱包的公钥和私钥
func (ws *Wallets) GetWallets() map[string]*Wallet {
	//打印公钥和私钥
	for address, wallet := range ws.Wallets {
		fmt.Printf("GetAddresses：address=%s\n", address)
		fmt.Printf("GetAddresses：wallet.PublicKey=%v\n", wallet.PublicKey)
		fmt.Printf("GetAddresses：wallet.PrivateKey=%v\n", wallet.PrivateKey)
	}
	return ws.Wallets
}

// GetWallet 这段代码是 `Wallets` 结构体的 `GetWallet` 方法，用于根据地址获取对应的钱包信息。
//下面是这个方法的功能和步骤解释：
//1. 使用给定的钱包地址 `address`，从 `Wallets` 实例中查找并获取对应的钱包。
//2. 返回找到的钱包对象。
//总的来说，这个方法的目的是根据地址获取钱包对象，以便在区块链交易中进行地址验证、签名等操作。
func (ws Wallets) GetWallet(address string) Wallet {
	return *ws.Wallets[address]
}

// LoadFromFile 这个方法是用于从文件中加载钱包数据到钱包集合（`Wallets`）中。下面是这个方法的功能解释：
//1. 构建钱包文件的路径，路径中包括了给定的 `nodeID`，用于确定钱包文件的名称。
//2. 检查钱包文件是否存在。如果文件不存在，返回错误，表示找不到钱包文件。
//3. 读取钱包文件的内容。如果读取文件出现错误，触发 Panic。
//4. 准备一个空的钱包集合实例，以便从文件中解码钱包数据并存储。
//5. 注册 elliptic.P256() 椭圆曲线参数，这是用于生成密钥对的一种算法。
//6. 创建一个新的解码器，用于从文件内容中解码钱包数据。
//7. 使用解码器将文件内容中的钱包数据解码到新的钱包集合实例中。
//8. 将新的钱包集合的钱包映射赋值给当前钱包集合实例的钱包映射，即更新当前钱包集合的钱包数据。
//9. 返回 nil 表示加载操作成功完成。
//总的来说，这个方法的目的是从文件中加载已存在的钱包数据到钱包集合中，以便可以在集合中管理和操作这些钱包的地址和密钥对。这是一个重要的功能，因为钱包数据用于签署和验证交易，从而确保区块链的安全和完整性。
func (ws *Wallets) LoadFromFile(nodeID string) error {
	walletFile := fmt.Sprintf(walletFile, nodeID)
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}

	fileContent, err := ioutil.ReadFile(walletFile)
	if err != nil {
		log.Panic(err)
	}

	var wallets Wallets
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)
	if err != nil {
		log.Panic(err)
	}

	ws.Wallets = wallets.Wallets

	return nil
}

// SaveToFile 这段代码是 `SaveToFile` 方法，属于 `Wallets` 结构体的方法，用于将钱包集合保存到文件中。
//以下是这个方法的功能和步骤解释：
//1. 创建一个字节缓冲区 `content` 用于存储编码后的钱包集合。
//2. 使用给定的 `nodeID` 构建钱包文件的文件名。
//3. 注册 `elliptic.P256()` 类型，以确保在编码和解码过程中正确处理椭圆曲线密钥。
//4. 创建一个新的 Gob 编码器，并使用它将钱包集合编码到 `content` 缓冲区中。
//5. 将编码后的数据写入文件，文件名为构建的钱包文件名，权限为 0644。
//总的来说，这个方法的目的是将钱包集合编码并保存到文件中，以便在之后重新加载时使用。钱包集合是用于管理多个钱包的数据结构，在区块链中用于存储和管理用户的密钥对和地址。
func (ws Wallets) SaveToFile(nodeID string) {
	var content bytes.Buffer
	walletFile := fmt.Sprintf(walletFile, nodeID)

	gob.Register(elliptic.P256())

	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws)
	if err != nil {
		log.Panic(err)
	}

	err = ioutil.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}

//签名函数
func (ws Wallets) Sign(address string, message []byte) (string, []byte, *big.Int, *big.Int, error) {
	fmt.Println("ws", ws)
	//找到Wallets中的指定地址
	wallet := ws.Wallets[address]
	//fmt.Printf("GetAddresses：wallet.PublicKey=%v\n", wallet.PublicKey)
	//fmt.Printf("GetAddresses：wallet.PrivateKey=%v\n", wallet.PrivateKey)
	//fmt.Println("wallet", wallet)
	// 对消息进行哈希
	//fmt.Println("message", message)
	hash := sha256.Sum256([]byte(message))
	//ecdsa.PrivateKey 转成 *PrivateKey
	//fmt.Println("hash", hash)
	// 使用私钥进行签名
	r, s, err := ecdsa.Sign(rand.Reader, &wallet.PrivateKey, hash[:])
	//fmt.Println("r", r)
	//fmt.Println("s", s)
	if err != nil {
		fmt.Println("Signature failed", err)
		return "", nil, nil, nil, err
	}
	// 签名成功
	//fmt.Printf("Signature (r, s): (%s, %s)\n", r, s)
	fmt.Println("Signature successful")
	return address, message, r, s, nil

}

// Verify 本地钱包验证签名
func (ws Wallets) Verify(address string, message string, r *big.Int, s *big.Int) bool {
	// 对消息进行哈希
	hash := sha256.Sum256([]byte(message))

	wallet := ws.Wallets[address]

	// 将 wallet.PublicKey 转换为 *ecdsa.PublicKey
	//var publicKey ecdsa.PublicKey
	publicKey.Curve = elliptic.P256() // 使用相应的椭圆曲线
	publicKey.X = new(big.Int).SetBytes(wallet.PublicKey[:32])
	publicKey.Y = new(big.Int).SetBytes(wallet.PublicKey[32:])

	//打印验证所需数据
	fmt.Println("address", address)
	fmt.Println("message", message)
	fmt.Println("r", r)
	fmt.Println("s", s)
	fmt.Println("publicKey", publicKey)
	fmt.Println("hash[:]", hash[:])
	// 使用公钥进行验证
	valid := ecdsa.Verify(&publicKey, hash[:], r, s)

	return valid
}

// SignByPrivateKey 签名函数
func SignByPrivateKey(command string) (string, []byte, *big.Int, *big.Int, error) {
	wallets, err := NewWallets(NodeIPAddress)
	if err != nil {
		log.Panic(err)
	}
	//fmt.Println("读取钱包成功", wallets)
	addresses := wallets.GetAddresses()
	commandByte := []byte(command)
	var r1, s1 *big.Int
	addresses[0], commandByte, r1, s1, _ = wallets.Sign(addresses[0], commandByte)
	//fmt.Println("签名成功")
	if wallets.Verify(addresses[0], command, r1, s1) {
		fmt.Println("Signature verification successful")
	} else {
		fmt.Println("signature verification failed")
	}
	return addresses[0], commandByte, r1, s1, nil
}

// VerifyByPublicKey 子节点通过公钥验证签名
func VerifyByPublicKey(publicKey ecdsa.PublicKey, message string, r *big.Int, s *big.Int) bool {
	// 对消息进行哈希
	hash := sha256.Sum256([]byte(message))

	// 使用公钥进行验证
	valid := ecdsa.Verify(&publicKey, hash[:], r, s)

	return valid
}
