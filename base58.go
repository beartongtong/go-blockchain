package main

import (
	"bytes"
	"math/big"
)

var b58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

// Base58Encode 这段代码是用于对 Base58 编码的数据进行解码的函数 `Base58Decode`。
//下面是这个函数的功能和步骤解释：
//1. 创建一个大整数 `result`，并初始化为 0，用于存储解码后的结果。
//2. 初始化变量 `zeroBytes` 为 0，用于统计输入数据中前导的 0 字节的数量。
//3. 循环遍历输入数据 `input` 中的每个字节 `b`，如果 `b` 是 0x00，则增加 `zeroBytes` 计数。
//4. 提取输入数据中去除前导 0 字节后的有效部分，保存在变量 `payload` 中。
//5. 遍历有效部分的每个字节 `b`，在 Base58 字母表 `b58Alphabet` 中查找字符 `b` 的索引位置 `charIndex`。
//6. 将 `result` 乘以 58，然后加上 `charIndex`，以将 Base58 编码转换为大整数形式。
//7. 将最终解码结果存储在变量 `decoded` 中，这是一个字节数组。
//8. 在 `decoded` 前面添加与前导 0 字节数量相等的字节 0x00，以恢复原始字节数组长度。
//9. 返回解码后的字节数组 `decoded`。
//总的来说，这个函数的目的是将使用 Base58 编码的字节数组解码为原始的字节数组，以便在区块链中进行地址验证等操作。它通过遍历输入数据的每个字节，
//根据 Base58 字母表进行解码，然后根据前导 0 字节的数量添加相应数量的 0 字节以恢复原始长度。
func Base58Encode(input []byte) []byte {
	var result []byte

	x := big.NewInt(0).SetBytes(input)

	base := big.NewInt(int64(len(b58Alphabet)))
	zero := big.NewInt(0)
	mod := &big.Int{}

	for x.Cmp(zero) != 0 {
		x.DivMod(x, base, mod)
		result = append(result, b58Alphabet[mod.Int64()])
	}

	ReverseBytes(result)
	for b := range input {
		if b == 0x00 {
			result = append([]byte{b58Alphabet[0]}, result...)
		} else {
			break
		}
	}

	return result
}

// Base58Decode 这段代码是用于对 Base58 编码的数据进行解码的函数 `Base58Decode`。
//下面是这个函数的功能和步骤解释：
//1. 创建一个大整数 `result`，并初始化为 0，用于存储解码后的结果。
//2. 初始化变量 `zeroBytes` 为 0，用于统计输入数据中前导的 0 字节的数量。
//3. 循环遍历输入数据 `input` 中的每个字节 `b`，如果 `b` 是 0x00，则增加 `zeroBytes` 计数。
//4. 提取输入数据中去除前导 0 字节后的有效部分，保存在变量 `payload` 中。
//5. 遍历有效部分的每个字节 `b`，在 Base58 字母表 `b58Alphabet` 中查找字符 `b` 的索引位置 `charIndex`。
//6. 将 `result` 乘以 58，然后加上 `charIndex`，以将 Base58 编码转换为大整数形式。
//7. 将最终解码结果存储在变量 `decoded` 中，这是一个字节数组。
//8. 在 `decoded` 前面添加与前导 0 字节数量相等的字节 0x00，以恢复原始字节数组长度。
//9. 返回解码后的字节数组 `decoded`。
//总的来说，这个函数的目的是将使用 Base58 编码的字节数组解码为原始的字节数组，以便在区块链中进行地址验证等操作。它通过遍历输入数据的每个字节，根据 Base58 字母表进行解码，然后根据前导 0 字节的数量添加相应数量的 0 字节以恢复原始长度。
func Base58Decode(input []byte) []byte {
	result := big.NewInt(0)
	zeroBytes := 0

	for b := range input {
		if b == 0x00 {
			zeroBytes++
		}
	}

	payload := input[zeroBytes:]
	for _, b := range payload {
		charIndex := bytes.IndexByte(b58Alphabet, b)
		result.Mul(result, big.NewInt(58))
		result.Add(result, big.NewInt(int64(charIndex)))
	}

	decoded := result.Bytes()
	decoded = append(bytes.Repeat([]byte{byte(0x00)}, zeroBytes), decoded...)

	return decoded
}
