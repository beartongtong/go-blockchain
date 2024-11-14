package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"log"

	"golang.org/x/crypto/ripemd160"
)

const version = byte(0x00)
const addressChecksumLen = 4

// Wallet stores private and public keys
type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

// NewWallet 这段代码是 `NewWallet` 函数，用于创建一个新的钱包。
//以下是这个函数的功能和步骤解释：
//1. 调用 `newKeyPair` 函数来生成一个新的密钥对，包括私钥和公钥。
//2. 使用生成的私钥和公钥创建一个新的钱包实例。
//3. 返回这个新钱包的指针。
//总的来说，这个函数的目的是生成一个新的密钥对并将其用于创建一个钱包实例。密钥对在加密货币中用于签署交易和验证身份。
func NewWallet() *Wallet {
	private, public := newKeyPair()
	wallet := Wallet{private, public}

	return &wallet
}

// GetAddress returns wallet address
func (w Wallet) GetAddress() []byte {
	pubKeyHash := HashPubKey(w.PublicKey)

	versionedPayload := append([]byte{version}, pubKeyHash...)
	checksum := checksum(versionedPayload)

	fullPayload := append(versionedPayload, checksum...)
	address := Base58Encode(fullPayload)

	return address
}

// HashPubKey hashes public key
func HashPubKey(pubKey []byte) []byte {
	publicSHA256 := sha256.Sum256(pubKey)

	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(publicSHA256[:])
	if err != nil {
		log.Panic(err)
	}
	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)

	return publicRIPEMD160
}

// ValidateAddress 这段代码定义了一个名为 `ValidateAddress` 的函数，用于验证区块链交易中的地址是否有效。
//1. `ValidateAddress` 函数接受一个字符串参数 `address`，表示待验证的地址。
//2. 首先，函数对输入的地址进行解码，将地址使用 `Base58Decode` 函数进行解码，得到公钥哈希（`pubKeyHash`）。
//3. 从 `pubKeyHash` 中获取实际的校验和（`actualChecksum`），校验和是公钥哈希的后几个字节，用于检测数据是否被篡改。
//4. 获取版本字节（`version`），它是公钥哈希的第一个字节，通常用于标识地址类型。
//5. 从 `pubKeyHash` 中移除版本字节和校验和部分，得到剩余的公钥哈希（`pubKeyHash`）。
//6. 计算目标校验和（`targetChecksum`），它是将版本字节和剩余的公钥哈希合并后再进行校验和计算。
//7. 最后，使用 `bytes.Compare` 函数比较实际校验和和目标校验和是否相等，如果相等则返回 `true`，表示地址有效，否则返回 `false`，表示地址无效。
//总之，这个函数用于验证区块链交易中的地址是否有效，通过比较校验和来检查地址的完整性和正确性。
func ValidateAddress(address string) bool {
	pubKeyHash := Base58Decode([]byte(address))
	actualChecksum := pubKeyHash[len(pubKeyHash)-addressChecksumLen:]
	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-addressChecksumLen]
	targetChecksum := checksum(append([]byte{version}, pubKeyHash...))

	return bytes.Compare(actualChecksum, targetChecksum) == 0
}

// Checksum generates a checksum for a public key
func checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])

	return secondSHA[:addressChecksumLen]
}

//这段代码用于生成一个新的ECDSA密钥对，包括私钥和相应的公钥。以下是代码的主要功能：
//1. 定义椭圆曲线，这里使用P-256曲线，即`elliptic.P256()`。
//2. 使用椭圆曲线和随机数生成函数（`rand.Reader`）来生成一个新的ECDSA私钥（`private`）。
//3. 如果生成私钥时出现错误，代码将引发日志记录（Panic）错误。
//4. 从生成的私钥中提取公钥的X和Y坐标，并将它们连接起来，形成公钥的字节切片。
//5. 返回生成的私钥（`private`）和公钥（`pubKey`）。
//这段代码用于生成密钥对，可以在区块链、加密货币等应用中用于签名和验证交易。私钥用于签署交易，而公钥用于验证签名
//。请注意，在实际应用中，密钥对的生成应该更加安全，并且需要妥善管理私钥，以确保其安全性。
func newKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	return *private, pubKey
}
