package main

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const protocol = "tcp"
const nodeVersion = 1
const commandLength = 36

var nodeAddress string
var miningAddress string
var myBestHeight int

//var knownNodes = []string{"localhost:3000"}
var knownNodes = []string{}
var knownShardingNodes = [][]string{} //分片节点信息列表、
//var knownShardingList = []string{}
var RelatedSharding = make(map[string]int) //关联分片信息 因为map没赋值的默认为0,为避免与分片0冲突，所以都+1，实际用时要-1
var blocksInTransit = [][]byte{}
var mempool = make(map[string]Transaction)
var firsthandleproposal = 0
var initversionflag = 0

//完成的提议
var completeproposal = make(map[string]*Proposal)

//var proposalpool = make(map[string]Proposal)
var BlockSyncnum = 0
var proposalpool = []Proposal{}
var QCpool = []QuorumCertificate{}
var frompool []string
var topool []string
var ProcessingProposalID string
var belongTo = ""      //节点所属分片
var NodeIP = ""        //节点IP,有:号
var NodeIPAddress = "" //节点IP,无:号，有空格
var belongToInt = -1   //节点所属分片

var startTime time.Time
var endTime time.Time

type addr struct {
	AddrList []string
}

type block struct {
	AddrFrom string
	Block    []byte
	ShardID  int
}

type getblocks struct {
	AddrFrom         string
	BlockchainHeight int
	belongTo         int
	ShardID          int
}

type getdata struct {
	AddrFrom string
	Type     string
	ID       []byte
	ShardID  int
}

type inv struct {
	AddrFrom string
	Type     string
	Items    [][]byte
	ShardID  int
}

type tx struct {
	AddFrom     string
	Transaction []byte
}

type Verzion struct {
	Version    int
	BestHeight int
	ShardID    int
	AddrFrom   string
}

//type Fragmentation struct {
//	knownNodes []string
//}

func commandToBytes(command string) []byte {
	var Tobytes [commandLength]byte

	for i, c := range command {
		Tobytes[i] = byte(c)
	}

	return Tobytes[:]
}

//func commandToBytes(command string) []byte {
//	var bytes []byte
//
//	for _, c := range command {
//		bytes = append(bytes, byte(c))
//	}
//
//	return bytes
//}

func bytesToCommand(bytes []byte) string {
	var command []byte

	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}

	return fmt.Sprintf("%s", command)
}

func extractCommand(request []byte) []byte {
	return request[:commandLength]
}

func requestBlocks() {
	//for _, node := range knownNodes {
	//	sendGetBlocks(node, myBestHeight)
	//}
}

func sendAddr(address string) {
	nodes := addr{knownNodes}
	nodes.AddrList = append(nodes.AddrList, nodeAddress)
	payload := gobEncode(nodes)
	request := append(commandToBytes("addr"), payload...)

	sendData(address, request)
}

func sendInit(address string) {
	nodes := addr{knownNodes}
	nodes.AddrList = append(nodes.AddrList, nodeAddress)
	payload := gobEncode(nodes)
	request := append(commandToBytes("addr"), payload...)

	sendData(address, request)
}

func sendBlock(addr string, b *Block, shardID int) {
	data := block{nodeAddress, b.Serialize(), shardID}
	payload := gobEncode(data)
	request := append(commandToBytes("block"), payload...)

	sendData(addr, request)
}

//这段代码定义了一个名为 `sendData` 的函数，用于将数据通过网络连接发送到指定的地址。
//以下是这个函数的功能和步骤解释：
//1. 使用给定的地址 `addr` 和协议 `protocol` 尝试与目标主机建立网络连接。
//2. 如果连接出现错误，表示目标主机不可用。在这种情况下，代码输出一条提示信息，更新已知节点列表 `knownNodes`，从中删除无法连接的地址。
//3. 如果连接成功建立，将数据从字节数组 `data` 通过网络连接传输到目标主机。
//4. 在完成数据传输后，关闭连接，释放资源。
//总的来说，这个函数用于发送数据到指定的网络地址，如果连接失败，则更新已知节点列表以排除不可用的节点。
//这在区块链网络中的节点之间进行通信时非常重要，以确保数据的传输和同步。
func sendData(addr string, data []byte) {
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		fmt.Printf("%s is not available\n", addr)
		var updatedNodes []string

		for _, node := range knownNodes {
			if node != addr {
				updatedNodes = append(updatedNodes, node)
			}
		}

		knownNodes = updatedNodes

		return
	}
	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}

func sendInv(address, kind string, items [][]byte, shardID int) {
	inventory := inv{nodeAddress, kind, items, shardID}
	payload := gobEncode(inventory)
	request := append(commandToBytes("inv"), payload...)

	sendData(address, request)
}

func sendGetBlocks(address string, BlockchainHeight int, shardID int) {
	payload := gobEncode(getblocks{nodeAddress, BlockchainHeight, belongToInt, shardID})
	request := append(commandToBytes("getblocks"), payload...)

	sendData(address, request)
}

//这段代码是用于向其他节点发送"getdata"请求的函数。在区块链网络中，节点之间可以请求特定类型（如区块或交易）的数据，
//以保持整个网络的数据同步。以下是这个函数的主要步骤和功能：
//1. 首先，它创建了一个 `getdata` 结构的实例，该结构包含了请求的详细信息，包括节点地址、请求类型（kind），以及要请求的数据的 ID。
//2. 然后，使用 `gobEncode` 函数将 `getdata` 结构编码成字节片（byte slice）。`gobEncode` 通常用于将 Go 中的结构序列化为字节数据，以便在网络上传输。
//3. 接下来，将 "getdata" 命令转换为字节并将其附加到请求中，以标识这是一个 "getdata" 请求。
//4. 最后，调用 `sendData` 函数，将请求数据发送给指定的节点地址（`address`）。
//总之，这段代码用于构建并发送 "getdata" 请求，以请求特定类型的数据（如区块或交易）以保持网络同步。
func sendGetData(address, kind string, id []byte, shardID int) {
	payload := gobEncode(getdata{nodeAddress, kind, id, shardID})
	request := append(commandToBytes("getdata"), payload...)

	sendData(address, request)
}

//这段代码是用于向其他节点发送交易信息的函数 `sendTx`。
//下面是这个函数的功能和步骤解释：
//1. 构造一个包含当前节点地址和序列化后的交易数据的结构 `tx`。
//2. 使用 `gobEncode` 函数对 `tx` 进行编码，将其转换为字节流。
//3. 构造一个请求数据，首先将命令类型 "tx" 转换为字节流，然后将上述编码后的数据追加到请求数据中。
//4. 调用 `sendData` 函数，将请求数据发送给目标节点的地址 `addr`。
//总的来说，这个函数的目的是将一个交易数据通过网络发送给其他节点，以便在区块链网络中传播交易信息。这有助于保持区块链的一致性和同步。
func sendTx(addr string, tnx *Transaction) {
	data := tx{nodeAddress, tnx.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("tx"), payload...)

	sendData(addr, request)
}

//这个函数用于发送版本信息给指定的节点。
//以下是这个函数的功能和步骤解释：
//1. 获取本节点区块链的最佳高度，即最新区块的高度，通过调用 `bc.GetBestHeight()` 方法。
//2. 创建一个版本信息的结构体，结构体包含当前节点的版本号、最佳高度和当前节点的地址。
//3. 使用 `gobEncode` 函数对版本信息结构体进行编码，得到序列化后的数据，即 payload。
//4. 使用 `commandToBytes` 函数将字符串命令 "version" 转换为字节数组。
//5. 将 payload 和命令字节数组合并为一个请求，即版本消息。
//6. 调用 `sendData` 函数，将版本消息发送给指定的地址。
//总的来说，这个函数的目的是向其他节点发送版本信息，用于在节点之间建立连接并同步区块链的信息。版本信息中包含当前节点的版本、最佳高度和地址等关键信息。
func sendVersion(addr string, bc *Blockchain, shardID int) {
	bestHeight := bc.GetBestHeight()
	fmt.Println("NodeIPAddress", NodeIPAddress)

	Verzion := Verzion{nodeVersion, bestHeight, shardID, NodeIPAddress}
	payload := gobEncode(Verzion)
	request := append(commandToBytes("version"), payload...)
	fmt.Println("sendData(addr, request):", addr)
	sendData(addr, request)
}

//测试数据结构
type Testdata struct {
	AddrFrom string
	Data     string
	From     string
	To       string
}

//发送测试数据
func sendTestdata(addr string, data string, from string, to string) {
	fmt.Println("sendTestdata to", addr, data)
	Testdata := Testdata{NodeIP, data, from, to}
	payload := gobEncode(Testdata)
	request := append(commandToBytes("sendTestdata"), payload...)
	fmt.Println("sendData(addr, request):", addr, data)
	sendData(addr, request)
}

type FragmentationData struct {
	AddrFrom           string
	KnownShardingNodes [][]string
	ShardID            int
	RelatedSharding    map[string]int
}

// SendknownShardingNodes
// 发送分片节点信息
func SendknownShardingNodes(addr string, ShardID int) {
	//element := []int{}
	//RelatedSharding = append(RelatedSharding, element)
	//RelatedSharding[0] = append(RelatedSharding[0], 1) //测试用 代表分片1为分片0的关联分片
	RelatedSharding["0-1"] = 3 //测试用 代表分片0和分片1的关联分片为分片2   值减1为关联分片的分片ID
	payload := gobEncode(FragmentationData{addr, knownShardingNodes, ShardID, RelatedSharding})
	fmt.Println("SendknownShardingNodes:", knownShardingNodes)
	request := append(commandToBytes("sendknownShardingNodes"), payload...)
	fmt.Println("sendData(addr, request):", addr)
	sendData(addr, request)
}

//preparePhaseData数据结构
type preparePhaseData struct {
	AddrFrom       string            //发送方地址
	Proposalvalue  Proposal          //提议
	QC             QuorumCertificate //QC证书
	ShardID        int               //分片ID
	TarGetShardID  int               //目标分片ID
	From           string            //发送方nodeID
	To             string            //接收方nodeID
	CrossShardFlag bool              //是否跨分片
}

// SendPrepareMsg 函数
// 发送prepare消息
func SendPrepareMsg(addr string, Proposalvalue Proposal, QC QuorumCertificate, from string, to string, shardID int, targetshardID int, CrossShardFlag bool) {
	//fmt.Println("Proposalvalue:", Proposalvalue)
	//fmt.Println("QC:", QC)
	//fmt.Println("QC.NodeSignatures[0].Tx:", QC.NodeSignatures[0].Tx)
	preparePhaseData := preparePhaseData{NodeIP, Proposalvalue, QC, shardID, targetshardID, from, to, CrossShardFlag}
	payload := gobEncode(preparePhaseData)
	request := append(commandToBytes("sendPrepareMsg"), payload...)
	fmt.Println("sendData(addr, request):", addr)
	sendData(addr, request)
}

//VoteMsg数据结构
type VoteMsg struct {
	AddrFrom      string
	BelongToInt   int
	Message       string
	Vote          Vote
	Proposalvalue Proposal
	ShardID       int //分片ID
	TarGetShardID int //目标分片ID
}

//sendVoteMsg函数
//发送vote消息
func sendVoteMsg(addr string, message string, vote Vote, proposal Proposal, shardID int, targetshardID int) {
	fmt.Println("sendVoteMsg")
	VoteMsg := VoteMsg{NodeIP, belongToInt, message, vote, proposal, shardID, targetshardID}
	payload := gobEncode(VoteMsg)
	request := append(commandToBytes("sendVoteMsg"), payload...)
	fmt.Println("sendData(addr, request):", addr)
	sendData(addr, request)

}

type BlockSyncData struct {
	AddrFrom    string
	BelongToInt int
}

//发送区块同步完成消息
func SendBlockSync(addr string) {
	payload := gobEncode(BlockSyncData{NodeIP, belongToInt})
	request := append(commandToBytes("sendBlockSync"), payload...)
	fmt.Println("sendData(addr, request):", addr)
	sendData(addr, request)
}

type CrossShardData struct {
	AddrFrom            string            //发送方地址
	Proposalvalue       Proposal          //提议
	QC                  QuorumCertificate //QC证书
	ShardID             int               //分片ID
	TarGetShardID       int               //目标分片ID
	From                string            //发送方nodeID
	To                  string            //接收方nodeID
	CrossShardFlag      bool              //是否跨分片
	RelatedShardingFlag bool              //是否关联分片
}

//发送跨分片交易数据
func SendCrossShardData(addr string, Proposalvalue Proposal, QC QuorumCertificate, from string, to string, shardID int, targetshardID int, CrossShardFlag bool, RelatedShardingFlag bool) {
	fmt.Println("SendCrossShardData to", addr, "Proposalvalue", Proposalvalue, "QC", QC, "shardID", shardID, "targetshardID", targetshardID, "from", from, "to", to, "CrossShardFlag", CrossShardFlag, "RelatedShardingFlag", RelatedShardingFlag)
	payload := gobEncode(CrossShardData{NodeIP, Proposalvalue, QC, shardID, targetshardID, from, to, CrossShardFlag, RelatedShardingFlag})
	request := append(commandToBytes("sendCrossShardData"), payload...)
	fmt.Println("sendData(addr, request):", addr)
	sendData(addr, request)
}

//commitPhaseData数据结构
type BalanceData struct {
	AddrFrom    string
	Address     string
	BelongToInt int
}

func SendBalanceMsg(addr string, address string) {
	fmt.Println("SendBalanceMsg to", addr, "belongToInt", belongToInt)
	payload := gobEncode(BalanceData{NodeIP, address, belongToInt})
	request := append(commandToBytes("sendBalanceMsg"), payload...)
	fmt.Println("sendData(addr, request):", addr)
	sendData(addr, request)
}

type TotalBalanceData struct {
	AddrFrom string
	Balance  int
}

func sendTotalBalanceMsg(addr string, balance int) {
	fmt.Println("sendTotalBalanceMsg balance", balance)
	payload := gobEncode(TotalBalanceData{NodeIP, balance})
	request := append(commandToBytes("sendTotalBalanceMsg"), payload...)
	fmt.Println("sendData(addr, request):", addr)
	sendData(addr, request)
}
func handleAddr(request []byte) {
	var buff bytes.Buffer
	var payload addr

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	knownNodes = append(knownNodes, payload.AddrList...)
	fmt.Printf("There are %d known nodes now!\n", len(knownNodes))
	requestBlocks()
}

func handleBlock(request []byte) {
	var buff bytes.Buffer
	var payload block

	bc := NewBlockchain(NodeIPAddress)
	defer bc.db.Close()
	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blockData := payload.Block
	block := DeserializeBlock(blockData)

	fmt.Println("Recevied a new block!")
	bc.AddBlock(block)
	UTXOSet := UTXOSet{bc}
	UTXOSet.Update(block)
	fmt.Printf("Added block %x\n", block.Hash)

	if len(blocksInTransit) > 0 {
		blockHash := blocksInTransit[0]
		sendGetData(payload.AddrFrom, "block", blockHash, payload.ShardID)

		blocksInTransit = blocksInTransit[1:]
	} else {
		fmt.Println("Block synchronization completed")

		//SendBlockSync(payload.AddrFrom)
		if NodeIPAddress != knownShardingNodes[belongToInt][0] {
			fmt.Println("Send synchronization request to leader")
			node := strings.Replace(knownShardingNodes[belongToInt][0], " ", ":", -1)
			SendBlockSync(node)
		}

	}
	//else {
	//	UTXOSet := UTXOSet{bc}
	//	UTXOSet.Reindex()
	//}
}

//这段代码是一个处理区块链网络消息（inventory）的函数。它的主要功能是解码接收到的消息，然后根据消息类型执行相应的操作。
//以下是该函数的主要步骤和功能：
//1. 首先，它创建一个`bytes.Buffer`类型的缓冲区，并将接收到的消息内容写入缓冲区，跳过命令部分。
//2. 然后，它使用`gob.NewDecoder`创建一个新的解码器，用于从缓冲区中解码数据。
//3. 使用解码器解码消息内容，并将解码的数据存储在`payload`变量中。
//4. 根据`payload`中的消息类型执行不同的操作。如果消息类型是 "block"，则表示接收到区块信息。它会更新`blocksInTransit`变量，
//然后向消息发送者请求具体的区块数据（通过`sendGetData`函数）。接着，它会更新`blocksInTransit`，将已请求的区块从待请求列表中移除。
//5. 如果消息类型是 "tx"，则表示接收到交易信息。它会检查内存池（mempool）中是否已经存在相同的交易，如果不存在，则向消息发送者请求具体的交易数据。
//总之，这段代码是用于处理区块链网络中传递的消息的一部分，它根据消息的内容和类型执行不同的操作，例如请求区块数据或交易数据，以确保区块链网络中的数据同步。
func handleInv(request []byte) {
	var buff bytes.Buffer
	var payload inv

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	fmt.Printf("Recevied inventory from %s\n", payload.AddrFrom)
	fmt.Printf("Recevied inventory with %d %s\n", len(payload.Items), payload.Type)

	if payload.Type == "block" {
		blocksInTransit = payload.Items

		blockHash := payload.Items[0]
		fmt.Println("sendGetData to", payload.AddrFrom, "block", blockHash, payload.ShardID)
		sendGetData(payload.AddrFrom, "block", blockHash, payload.ShardID)

		newInTransit := [][]byte{}
		for _, b := range blocksInTransit {
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}
		blocksInTransit = newInTransit
	}

	if payload.Type == "tx" {
		txID := payload.Items[0]

		if mempool[hex.EncodeToString(txID)].ID == nil {
			sendGetData(payload.AddrFrom, "tx", txID, payload.ShardID)
		}
	}
}

//这段代码处理 `getblocks` 命令，其中 `getblocks` 是一个结构体类型。让我解释一下代码的功能：
//1. `handleGetBlocks` 函数接收两个参数：`request` 和 `bc`。
//- `request` 是接收到的命令字节，其中可能包含了一些数据。
//- `bc` 是一个指向 `Blockchain` 结构体的指针，表示区块链。
//2. 创建一个 `bytes.Buffer` 类型的变量 `buff`，用于缓存数据。
//3. 创建一个名为 `payload` 的变量，其类型是 `getblocks` 结构体。
//4. 将 `request` 切片中从 `commandLength` 位置开始的数据写入到缓冲区 `buff` 中。
//5. 使用 `gob.NewDecoder` 创建一个解码器 `dec`，并将其与缓冲区 `buff` 关联。
//6. 使用 `dec.Decode` 将缓冲区中的数据解码到 `payload` 变量中。如果解码出现错误，则记录日志并触发 panic。
//7. 调用 `bc.GetBlockHashes()` 获取区块链中所有区块的哈希值，并将这些哈希值保存在 `blocks` 变量中。
//8. 调用 `sendInv` 函数，向 `payload.AddrFrom` 地址发送 `block` 类型的 `inv` 消息，携带区块哈希值列表 `blocks`。
//总体来说，这段代码的作用是处理 `getblocks` 命令，解码其中的数据，然后通过 `sendInv` 函数向指定地址发送区块哈希值列表。
//在区块链网络中，`getblocks` 命令用于请求其他节点发送它们所拥有的区块的哈希值列表。
func handleGetBlocks(request []byte) {
	var buff bytes.Buffer
	var payload getblocks

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	//根据payload.belongTo判断来者要求的是哪个分片的数据
	if payload.ShardID == belongToInt {
		fmt.Println("---------------------")
		fmt.Println("Receive from", payload.AddrFrom, "的getblocks请求 1")
		fmt.Println("payload.belongTo:", payload.belongTo)
		bc := NewBlockchain(NodeIPAddress)
		//区块高度从0开始，要+1
		fmt.Println("payload.BlockchainHeight:", payload.BlockchainHeight+1)
		blocks := bc.GetBlockHashes()
		defer bc.db.Close()
		fmt.Println("len(blocks):", len(blocks))
		//截取blocks的前payload.BlockchainHeight个区块
		if len(blocks) > payload.BlockchainHeight {
			blocks = blocks[:len(blocks)-(payload.BlockchainHeight+1)]
		}
		//fmt.Println("截取后的len(blocks):", len(blocks))
		//sendInv(payload.AddrFrom, "block", [][]byte{blocks[0]})
		sendInv(payload.AddrFrom, "block", blocks, belongToInt)

		fmt.Println("---------------------")
	} else {
		fmt.Println("---------------------")
		fmt.Println("Receive from", payload.AddrFrom, "的getblocks请求 2")
		fmt.Println("payload.ShardID:", payload.ShardID)
		//defer bc.db.Close()
		newbc := NewBlockchain(knownShardingNodes[payload.ShardID][0])
		defer newbc.db.Close()
		//区块高度从0开始，要+1
		fmt.Println("payload.BlockchainHeight:", payload.BlockchainHeight+1)
		blocks := newbc.GetBlockHashes()

		fmt.Println("len(blocks):", len(blocks))
		//截取blocks的前payload.BlockchainHeight个区块
		if len(blocks) > payload.BlockchainHeight {
			blocks = blocks[:len(blocks)-(payload.BlockchainHeight+1)]
		}
		//fmt.Println("截取后的len(blocks):", len(blocks))
		//sendInv(payload.AddrFrom, "block", [][]byte{blocks[0]})
		sendInv(payload.AddrFrom, "block", blocks, payload.ShardID)

		fmt.Println("---------------------")

	}

}

//这段代码处理了来自其他节点的 "getdata" 请求。在区块链网络中，节点可以向其他节点发送 "getdata" 请求，以请求特定类型（如区块或交易）的数据。
//以下是这个函数的主要步骤和功能：
//1. 首先，它创建了一个 `getdata` 结构的实例 `payload`，并从请求中解码出这个结构。`getdata` 结构包含有关请求的详细信息，
//包括请求类型（"block" 或 "tx"）、数据的 ID 和发送请求的节点地址。
//2. 接下来，根据请求的类型，它执行以下操作：
//- 如果请求的类型是 "block"，则调用区块链的 `GetBlock` 函数，根据传入的区块 ID 获取相应的区块。然后，使用 `sendBlock` 函数将该区块发送回请求的节点。
//- 如果请求的类型是 "tx"，则从内存池中查找相应的交易（使用交易 ID），然后使用 `sendTx` 函数将该交易发送回请求的节点。
//总之，这段代码用于处理 "getdata" 请求，根据请求的类型发送相应的数据（区块或交易）给请求的节点，以满足区块链网络中节点之间的数据同步需求。
func handleGetData(request []byte) {
	var buff bytes.Buffer
	var payload getdata

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	fmt.Println("Receive from", payload.AddrFrom, "的getdata请求")
	fmt.Println("handleGetData")
	fmt.Println("payload.Type", payload.Type)
	var bc *Blockchain
	if payload.ShardID != belongToInt {
		bc = NewBlockchain(knownShardingNodes[payload.ShardID][0])
	} else {
		bc = NewBlockchain(NodeIPAddress)
	}
	defer bc.db.Close()
	if payload.Type == "block" {
		fmt.Println("block, err := bc.GetBlock([]byte(payload.ID)) start")

		block, err := bc.GetBlock([]byte(payload.ID))
		fmt.Println("block, err := bc.GetBlock([]byte(payload.ID)) end")
		if err != nil {
			fmt.Println("err != nil:", err)
			return
		}
		fmt.Println("payload.AddrFrom:", payload.AddrFrom)
		fmt.Println("sendBlock(payload.AddrFrom, &block)")
		sendBlock(payload.AddrFrom, &block, payload.ShardID)
	}

	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)
		tx := mempool[txID]

		sendTx(payload.AddrFrom, &tx)
		//delete(mempool, txID)
	}
	//defer bc.db.Close()
	//bc.db.Close()

}

//这段代码是一个用于处理交易的函数 `handleTx`，它根据接收到的交易数据执行不同的操作，包括将交易放入内存池（mempool）、进行挖矿以创建新区块等。
//函数解释如下：
//1. 从收到的请求数据中解析出交易数据。
//2. 反序列化交易数据，得到交易对象 `tx`。
//3. 将交易添加到内存池 `mempool`，使用交易 ID 作为键。
//4. 如果当前节点是网络中的主节点（knownNodes[0]），则向其他节点广播交易的库存信息（inv）。
//5. 否则，如果内存池中有足够的交易且矿工地址不为空，开始挖矿操作。
//- 遍历内存池中的交易，验证每个交易，将有效交易加入到 `txs` 切片。
//- 创建一个 coinbase 交易并添加到 `txs` 切片。
//- 使用 `bc.MineBlock` 函数进行挖矿，创建新区块。
//- 更新 UTXO 集合，重新索引。
//- 删除已挖矿的交易。
//- 向其他节点广播新挖矿的区块信息。
//- 如果内存池中还有其他交易，则继续挖矿，直到内存池为空或没有足够的交易。
//总的来说，这个函数用于处理交易，包括将交易添加到内存池、进行挖矿并创建新区块，以及向其他节点广播相关信息。它是区块链网络中的一个关键部分，用于维护交易的流动和区块的生成。
func handleTx(request []byte) {
	var buff bytes.Buffer
	var payload tx
	fmt.Println("handleTx")
	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	bc := NewBlockchain(NodeIPAddress)
	fmt.Println("payload.AddFrom:", payload.AddFrom)
	txData := payload.Transaction
	tx := DeserializeTransaction(txData)
	mempool[hex.EncodeToString(tx.ID)] = tx
	fmt.Println("len(mempool):", len(mempool))
	fmt.Println("len(miningAddress):", len(miningAddress))
	if nodeAddress == knownNodes[0] {
		for _, node := range knownNodes {
			if node != nodeAddress && node != payload.AddFrom {
				sendInv(node, "tx", [][]byte{tx.ID}, belongToInt)
			}
		}
	} else {
		if len(mempool) >= 1 && len(miningAddress) > 0 {
		MineTransactions:
			var txs []*Transaction

			for id := range mempool {
				tx := mempool[id]
				if bc.VerifyTransaction(&tx) {
					txs = append(txs, &tx)
				}
			}

			if len(txs) == 0 {
				fmt.Println("All transactions are invalid! Waiting for new ones...")
				return
			}
			fmt.Println("len(txs):", len(txs))

			cbTx := NewCoinbaseTX(miningAddress, "")
			txs = append(txs, cbTx)

			fmt.Println("newBlock := bc.MineBlock(txs)")
			newBlock := bc.MineBlock(txs)
			fmt.Println("UTXOSet := UTXOSet{bc}")
			UTXOSet := UTXOSet{bc}
			fmt.Println("UTXOSet.Reindex()")
			UTXOSet.Reindex()

			fmt.Println("New block is mined!")

			for _, tx := range txs {
				txID := hex.EncodeToString(tx.ID)
				delete(mempool, txID)
			}

			for _, node := range knownNodes {
				if node != nodeAddress {
					sendInv(node, "block", [][]byte{newBlock.Hash}, belongToInt)
				}
			}

			if len(mempool) > 0 {
				goto MineTransactions
			}
		}
		//if len(miningAddress) <= 0 {
		//	fmt.Println("Not enough miningAddress! Waiting for new ones...")
		//}
		//if len(mempool) <= 1 {
		//	fmt.Println("Not enough transactions! Waiting for new ones...")
		//}
	}
}

//这个函数是处理版本消息的逻辑，版本消息在区块链网络中用于节点之间的握手和信息交换。
//以下是这个函数的功能和步骤解释：
//1. 创建一个新的缓冲区 `buff`，用于存储消息的有效负载。
//2. 将消息的有效负载写入缓冲区 `buff`，即去掉消息的命令部分，以便进行后续解码。
//3. 创建一个新的 `version` 变量，用于存储解码后的版本消息。
//4. 使用 `gob.NewDecoder` 创建一个解码器 `dec`，用于将缓冲区中的数据解码为版本消息。
//5. 将缓冲区中的数据通过解码器进行解码，将解码后的版本消息存储在 `version` 变量中。
//6. 如果在解码过程中出现错误，即无法将数据解码为版本消息，将触发 Panic。
//7. 获取本地区块链的最佳高度 `myBestHeight`。
//8. 获取收到版本消息中的对方节点的最佳高度 `foreignerBestHeight`。
//9. 根据最佳高度的比较，判断是否需要请求区块数据：
//- 如果本地节点的最佳高度小于对方节点的最佳高度，向对方节点发送 `getblocks` 请求以获取缺失的区块数据。
//- 如果本地节点的最佳高度大于对方节点的最佳高度，向对方节点发送版本消息以告知本地节点的最新信息。
//10. 如果对方节点不在已知节点列表中，将其添加到已知节点列表中。
//总的来说，这个函数用于处理版本消息，进行节点间的握手和信息交换，以保持区块链网络的同步和一致性。
func handleVersion(request []byte) {
	var buff bytes.Buffer
	var payload Verzion

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	fmt.Println("---------------")
	fmt.Println("handleVersion")
	fmt.Println("nodeIPAddress:", NodeIPAddress)
	fmt.Println("nodeIP:", NodeIP)
	fmt.Println("Receive from:", payload.AddrFrom)
	fmt.Println("payload.ShardID:", payload.ShardID)

	if payload.ShardID != belongToInt {
		newbc := NewBlockchain(knownShardingNodes[payload.ShardID][0])
		defer newbc.db.Close()
		foreignerBestHeight := payload.BestHeight
		myBestHeight = newbc.GetBestHeight()
		node := strings.Replace(payload.AddrFrom, " ", ":", -1)
		if myBestHeight < foreignerBestHeight {
			sendGetBlocks(node, myBestHeight, payload.ShardID)
		} else if myBestHeight > foreignerBestHeight {
			sendVersion(node, newbc, payload.ShardID)
		}
	} else {
		bc := NewBlockchain(NodeIPAddress)
		defer bc.db.Close()
		foreignerBestHeight := payload.BestHeight
		myBestHeight = bc.GetBestHeight()
		fmt.Println("myBestHeight:", myBestHeight)
		fmt.Println("foreignerBestHeight:", foreignerBestHeight)
		if myBestHeight < foreignerBestHeight {
			fmt.Println("myBestHeight < foreignerBestHeight")
			node := strings.Replace(payload.AddrFrom, " ", ":", -1)
			fmt.Println("sendGetBlocks to ", node)
			sendGetBlocks(node, myBestHeight, payload.ShardID)
		} else if myBestHeight > foreignerBestHeight {
			fmt.Println("myBestHeight > foreignerBestHeight")
			fmt.Println("myBestHeight:", myBestHeight)
			fmt.Println("foreignerBestHeight:", foreignerBestHeight)
			node := strings.Replace(payload.AddrFrom, " ", ":", -1)
			sendVersion(node, bc, payload.ShardID)
		}
		fmt.Println("---------------")
		// sendAddr(payload.AddrFrom)
		if !nodeIsKnown(payload.AddrFrom) {
			knownNodes = append(knownNodes, payload.AddrFrom)
		}
	}

}

func handleSendTestdata(request []byte) {
	var buff bytes.Buffer
	var payload Testdata

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	fmt.Println("handleSendTestdata")
	fmt.Println(payload.Data)
	substrings := strings.Split(payload.Data, " ")
	//fmt.Println("createProposal")
	fmt.Println("Command:", payload.Data)
	//fmt.Println("knownShardingNodes[belongToInt][0]", knownShardingNodes[belongToInt][0])
	leaderID := strings.Replace(knownShardingNodes[belongToInt][0], " ", ":", -1)
	//fmt.Println("leaderID", leaderID)
	node := Node{
		ID:             NodeIPAddress,
		Proposals:      []*Proposal{},
		LastProposalID: "",
		Mutex:          sync.Mutex{},
	}
	fmt.Println("node", node)
	bc := NewBlockchain(NodeIPAddress)
	defer bc.db.Close()
	switch substrings[1] {
	case "send":
		from := substrings[3]
		to := substrings[5]
		amount, err := strconv.Atoi(substrings[7])
		nodeID := NodeIPAddress
		mineNow := false
		fmt.Println("nodeID", nodeID)
		if substrings[8] == "mine" {
			mineNow = true
			fmt.Println("mine now：", mineNow)
		}
		//fmt.Println("mineNow", mineNow)
		if !ValidateAddress(from) {
			log.Panic("ERROR: Sender address is not valid")
		}
		//fmt.Println("from", from)
		if !ValidateAddress(to) {
			log.Panic("ERROR: Recipient address is not valid")
		}

		UTXOSet := UTXOSet{bc}
		//fmt.Println("UTXOSet", UTXOSet)
		//defer bc.db.Close()
		//if bc != nil {
		//	bc.db.Close()
		//}
		wallets, err := NewWallets(nodeID)
		//fmt.Println("wallets", wallets)
		if err != nil {
			log.Panic(err)
		}
		wallet := wallets.GetWallet(from)
		fmt.Println("wallet", wallet)
		fmt.Println("to", to)
		fmt.Println("amount", amount)
		tx := NewUTXOTransaction(&wallet, to, amount, &UTXOSet)
		//fmt.Println("tx", tx)
		if tx != nil {
			preparePhase(leaderID, 0, &node, payload.Data, from, tx, payload.From, payload.To)
		} else {
			fmt.Println("tx is nil,可能是钱不够")
		}
	case "DistributeRewards":
		fmt.Println("DistributeRewards")
		//nodeID := NodeIPAddress
		//bc := NewBlockchain(nodeID)
		fmt.Println("1")
		address := substrings[3]
		amount, err := strconv.Atoi(substrings[4])

		if err != nil {
			log.Panic(err)
		}
		if !ValidateAddress(address) {
			fmt.Println("ERROR: Address is not valid")
		} else {
			cbtx := NewCoinbaseTX2(address, payload.Data, amount)
			fmt.Println("2")
			sourcetxs := []*Transaction{cbtx}
			var newSourceBlock *Block

			sourceUTXOSet := UTXOSet{bc} //发起分片

			newSourceBlock = bc.commitTransaction(sourcetxs, payload.Data)
			fmt.Println("3")
			sourceUTXOSet.Update(newSourceBlock)
			fmt.Println("4")
			fmt.Printf("----Added block %x\n", newSourceBlock.Hash)
		}

		fmt.Println("区块链上链成功!")
		for _, node := range knownShardingNodes[belongToInt] {
			if node == NodeIPAddress {
				fmt.Println("数据以上链，不需要给自己发同步信息")

			} else {

				node = strings.Replace(node, " ", ":", -1)
				//fmt.Println("给子节点更新区块：", node)
				BlockSyncnum = 0
				//广播区块
				sendVersion(node, bc, belongToInt)
			}
		}
		fmt.Println("5")
	case "getbalance":
		fmt.Println("getbalance")
		fmt.Println("AddrFrom:", payload.AddrFrom)
		address := substrings[3]
		if !ValidateAddress(address) {
			log.Panic("ERROR: Address is not valid")
		}
		UTXOSet := UTXOSet{bc}

		//defer bc.db.Close()

		balance := 0
		pubKeyHash := Base58Decode([]byte(address))
		pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]

		UTXOs := UTXOSet.FindUTXO(pubKeyHash)

		for _, out := range UTXOs {
			balance += out.Value
		}
		fmt.Println("Balance:", balance)
		//totalBalance = totalBalance + balance
		//给所有分片领导者发消息
		//遍历knownShardingNodes
		for index, knownShardingNode := range knownShardingNodes {
			if index != belongToInt {
				//遍历knownShardingNode
				fmt.Println("向领导者：", knownShardingNode[0], "发送消息")
				sendNodeID := strings.Replace(knownShardingNode[0], " ", ":", -1)
				//给分片领导者发消息
				SendBalanceMsg(sendNodeID, address)
			} else {
				//如果自己不是领导者
				if NodeIPAddress != knownShardingNodes[belongToInt][0] {
					//向领导者发送消息
					sendNodeID := strings.Replace(knownShardingNodes[belongToInt][0], " ", ":", -1)
					sendTotalBalanceMsg(sendNodeID, balance)
				}
			}
		}
		//data := map[string]interface{}{
		//	"command":            "StatisticalBalance",
		//	"balance":            balance,
		//	"wallet":             "",
		//	"IP":                 "",
		//	"port":               "",
		//	"to":                 "",
		//	"from":               "",
		//	"knownShardingNodes": knownShardingNodes,
		//}
		//// 将数据编码为JSON
		//jsonData, err := json.Marshal(data)
		//if err != nil {
		//	log.Fatalf("Error occurred during JSON marshalling: %v", err)
		//}
		//
		//// 发送POST请求
		//resp, err := http.Post("http://192.168.254.129:8088/post", "application/json", bytes.NewBuffer(jsonData))
		//if err != nil {
		//	log.Fatalf("An Error Occurred: %v", err)
		//}
		////defer resp.Body.Close()
		////fmt.Println("Response status:", resp.Status)
		//defer resp.Body.Close()
		//// 打印响应状态
		//log.Println("Response status:", resp.Status)
		////SendBalanceMsg(payload.AddrFrom, address, balance)
	}

}
func handleSendknownShardingNodes(request []byte) {
	var buff bytes.Buffer
	var payload FragmentationData

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	fmt.Println(payload)
	knownShardingNodes = payload.KnownShardingNodes
	RelatedSharding = payload.RelatedSharding
	if belongToInt == -1 {
		//初始化节点分片ID
		belongToInt = payload.ShardID //将belongToInt赋值为分片ID
		belongTo = strconv.Itoa(belongToInt)
		fmt.Println("Update node list：", knownShardingNodes)
		fmt.Println("Update association list：", RelatedSharding)
		fmt.Println("This node belongs to sharding:", belongToInt)
		fmt.Println("current node:", NodeIPAddress)
		//向所有人发送节点信息
		for i, _ := range knownShardingNodes {
			for _, nodeAddress := range knownShardingNodes[i] {
				//如果节点是自己则不发送,防止自己给自己发送节点信息，同步两次消息
				if nodeAddress != NodeIPAddress {
					targetNodeAddress := strings.Replace(nodeAddress, " ", ":", -1)
					fmt.Println("向分片", i, "-", targetNodeAddress, "发送节点信息")
					SendknownShardingNodes(targetNodeAddress, payload.ShardID)
				}
			}
		}
	} else {
		fmt.Println("更新节点列表：", knownShardingNodes)
	}

	//:号替换成空格
	nodeAddress := strings.Replace(payload.AddrFrom, ":", " ", -1)

	if nodeAddress != knownShardingNodes[belongToInt][0] && initversionflag == 0 { //如果不是领导者节点
		fmt.Println("sendVersion to", knownShardingNodes[belongToInt][0])
		bc := NewBlockchain(NodeIPAddress)
		defer bc.db.Close()
		ShardLeaderIP := strings.Replace(knownShardingNodes[belongToInt][0], " ", ":", -1)
		sendVersion(ShardLeaderIP, bc, belongToInt) //向领导者节点发送版本消息
		initversionflag = 1
		//SendknownShardingNodes(ShardLeaderIP, payload.ShardID) //向领导者节点发送分片节点信息
	}
}
func handleSendPrepareMsg(request []byte) {
	var buff bytes.Buffer
	var payload preparePhaseData
	UsedTxFlag := 0
	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	//if NodeIPAddress == knownShardingNodes[payload.ShardID][0] {
	if amILeader(NodeIPAddress, payload.TarGetShardID) {
		fmt.Println("这是领导者节点", belongToInt, "，需要给其他子节点发消息")
		fmt.Println("NodeIP:", NodeIP, "接收到来自", payload.AddrFrom, "的PrepareMsg消息")
		for i, tx := range payload.QC.NodeSignatures[0].Tx.Vin {
			if UsedTxId[hex.EncodeToString(tx.Txid)] != nil {
				fmt.Println("第", i, "个交易已经被使用过了")
				UsedTxFlag = 1
			}
		}
		//UsedTxFlag = 0
		if UsedTxFlag == 0 {

			fmt.Println("UsedTxId", UsedTxId[hex.EncodeToString(payload.QC.NodeSignatures[0].Tx.Vin[0].Txid)])
			UsedTxId[hex.EncodeToString(payload.QC.NodeSignatures[0].Tx.Vin[0].Txid)] = payload.QC.NodeSignatures[0].Tx.Vin[0].Txid
			//将Proposalvalue存入proposalpool
			proposalpool = append(proposalpool, payload.Proposalvalue)
			//将QC存入QCpool
			QCpool = append(QCpool, payload.QC)
			//将from存入frompool
			frompool = append(frompool, payload.From)
			//将to存入topool
			topool = append(topool, payload.To)

			fmt.Println("-------------------------------")
			fmt.Println("目前提议池中已有：", len(proposalpool), "个提议")
			fmt.Println("最新提议是：", proposalpool[len(proposalpool)-1].ID)
			fmt.Println("当前处理提议：", ProcessingProposalID)
			fmt.Println("-------------------------------")
			if len(proposalpool) == 1 {
				startTime = time.Now()
				ProcessingProposalID = proposalpool[0].ID
				fmt.Println("开始处理提议：", ProcessingProposalID)
				go handleProposal(payload.ShardID, payload.TarGetShardID, payload.QC, payload.Proposalvalue, payload.From, payload.To)
				//if VerifyByPublicKey(payload.QC.NodeSignatures[0].PublicKey, payload.QC.Message.Value, payload.QC.NodeSignatures[0].R, payload.QC.NodeSignatures[0].S) {
				//	fmt.Println("领导者验证通过")
				//	var wallet string
				//	var r1, s1 *big.Int
				//	//返回签名给领导者节点
				//	wallet, _, r1, s1, _ = SignByPrivateKey(payload.QC.Message.Value)
				//	var vote Vote
				//	vote.Votetype = "agree"
				//	vote.NodeID = NodeIPAddress
				//	vote.Addresss = wallet
				//	vote.R = r1
				//	vote.S = s1
				//	vote.PublicKey = publicKey
				//	vote.Tx = payload.QC.NodeSignatures[0].Tx
				//	requiredAgree := (len(knownShardingNodes[belongToInt]) / 2) + 1 //只需要发起分片和目标分片的领导者同意即可
				//	createOrUpdateVoteCollector(vote, payload.QC.Message.Value, bc, payload.Proposalvalue, requiredAgree, payload.ShardID, payload.TarGetShardID)
				//	go handleProposal(payload.ShardID, payload.TarGetShardID, payload.QC, payload.Proposalvalue, payload.From, payload.To)
				//	//sendVoteMsg(payload.AddrFrom, payload.QC.Message.Value, vote, payload.Proposalvalue, payload.ShardID, payload.TarGetShardID)
				//} else {
				//	fmt.Println("验证不通过")
				//	var vote Vote
				//	vote.Votetype = "disagree"
				//}
			}

		} else {
			fmt.Println("该提议的交易已被提交过")
			sendTestdata(payload.AddrFrom, payload.Proposalvalue.Value, payload.From, payload.To)
		}

	} else {
		fmt.Println("这是非领导者节点，不需要给其他子节点发消息，开始验证提议信息，准备投票")
		fmt.Println("NodeIP:", NodeIP, "接收到来自", payload.AddrFrom, "的PrepareMsg消息")
		//fmt.Println("payload.QC.NodeSignatures[0].PublicKey", payload.QC.NodeSignatures[0].PublicKey)
		fmt.Println("提议内容：", payload.QC.Message.Value)
		//fmt.Println("payload.QC.NodeSignatures[0].R", payload.QC.NodeSignatures[0].R)
		//fmt.Println("payload.QC.NodeSignatures[0].S", payload.QC.NodeSignatures[0].S)
		//验证提议信息
		if VerifyByPublicKey(payload.QC.NodeSignatures[0].PublicKey, payload.QC.Message.Value, payload.QC.NodeSignatures[0].R, payload.QC.NodeSignatures[0].S) {
			fmt.Println("验证通过,返回签名给领导者节点")
			var wallet string
			var r1, s1 *big.Int
			//返回签名给领导者节点
			wallet, _, r1, s1, _ = SignByPrivateKey(payload.QC.Message.Value)
			var vote Vote
			vote.Votetype = "agree"
			vote.NodeID = NodeIPAddress
			vote.Addresss = wallet
			vote.R = r1
			vote.S = s1
			vote.PublicKey = publicKey
			vote.Tx = payload.QC.NodeSignatures[0].Tx
			sendVoteMsg(payload.AddrFrom, payload.QC.Message.Value, vote, payload.Proposalvalue, payload.ShardID, payload.TarGetShardID)
		} else {
			fmt.Println("验证不通过")
			var vote Vote
			vote.Votetype = "disagree"
		}
	}

}

func handleProposal(ShardID int, TarGetShardID int, QC QuorumCertificate, proposal Proposal, from string, to string) {
	fmt.Println("handleProposal")
	fmt.Println("ProcessingProposalID", ProcessingProposalID)
	fmt.Println("proposal.ID", proposal.ID)
	if ProcessingProposalID == proposal.ID {
		fmt.Println("处理提议：", proposal.ID)
		targetShardID := -1
		//遍历knownShardingNodes[]数组
		fmt.Println("knownShardingNodes:", knownShardingNodes)
		for i, _ := range knownShardingNodes {
			if targetShardID >= 0 {
				break
			} else {
				for _, node := range knownShardingNodes[i] {
					fmt.Println("node:", node)
					if node == to {
						targetShardID = i
						break
					}
				}
			}
		}
		if targetShardID != -1 {
			if targetShardID == belongToInt {
				fmt.Println("非跨分片交易")
				for _, node := range knownShardingNodes[ShardID] {
					if node != NodeIPAddress {
						node = strings.Replace(node, " ", ":", -1)
						//给分片中除自己外的其他人发送消息
						SendPrepareMsg(node, proposal, QC, from, to, ShardID, ShardID, false)
					} else {
						fmt.Println("这是领导者自身节点，不需要发消息")
						if VerifyByPublicKey(QC.NodeSignatures[0].PublicKey, QC.Message.Value, QC.NodeSignatures[0].R, QC.NodeSignatures[0].S) {
							fmt.Println("验证通过,返回签名给领导者节点")
							var wallet string
							var r1, s1 *big.Int
							//返回签名给领导者节点
							wallet, _, r1, s1, _ = SignByPrivateKey(QC.Message.Value)
							var vote Vote
							vote.Votetype = "agree"
							vote.NodeID = NodeIPAddress
							vote.Addresss = wallet
							vote.R = r1
							vote.S = s1
							vote.PublicKey = publicKey
							vote.Tx = QC.NodeSignatures[0].Tx
							node = strings.Replace(node, " ", ":", -1)
							sendVoteMsg(node, QC.Message.Value, vote, proposal, ShardID, TarGetShardID)
						} else {
							fmt.Println("验证不通过")
							var vote Vote
							vote.Votetype = "disagree"
						}
					}
				}
			} else {
				fmt.Println("跨分片交易")
				fmt.Println("发起分片ID：", belongToInt)
				fmt.Println("目标分片ID：", targetShardID)
				relatedShardingflag := false
				crossShardingflag := true

				//for _, relatedShardID := range RelatedSharding[ShardID] {
				//	if relatedShardID == TarGetShardID {
				//		relatedShardingflag = true
				//		break
				//	}
				//}
				result := strconv.Itoa(belongToInt) + "-" + strconv.Itoa(targetShardID)
				if RelatedSharding[result] > 0 {
					relatedShardingflag = true
					addrIP := strings.Replace(knownShardingNodes[RelatedSharding[result]-1][0], " ", ":", -1)
					SendCrossShardData(addrIP, proposal, QC, from, to, ShardID, targetShardID, crossShardingflag, relatedShardingflag)
				}

				if relatedShardingflag == false {
					//广播准备消息给发起分片领导者
					addrIP := strings.Replace(knownShardingNodes[ShardID][0], " ", ":", -1)
					SendCrossShardData(addrIP, proposal, QC, from, to, ShardID, targetShardID, crossShardingflag, relatedShardingflag)

					//广播准备消息给目标分片领导者
					addrIP = strings.Replace(knownShardingNodes[targetShardID][0], " ", ":", -1)
					SendCrossShardData(addrIP, proposal, QC, from, to, ShardID, targetShardID, crossShardingflag, relatedShardingflag)
					////目标分片不是关联分片,向两个分片所有节点发送投票信息
					////向发起分片所有节点发送消息
					//for _, node := range knownShardingNodes[ShardID] {
					//	if node != knownShardingNodes[ShardID][0] {
					//		node = strings.Replace(node, " ", ":", -1)
					//		//给分片中除自己外的其他人发送消息
					//		SendPrepareMsg(node, proposal, QC, from, to, ShardID, ShardID, false)
					//	} else {
					//		fmt.Println("这是领导者自身节点，不需要发消息")
					//	}
					//}
					////向目标分片所有节点发送消息
					//for _, node := range knownShardingNodes[targetShardID] {
					//	if node != knownShardingNodes[targetShardID][0] {
					//		node = strings.Replace(node, " ", ":", -1)
					//		//给分片中除自己外的其他人发送消息
					//		SendPrepareMsg(node, proposal, QC, from, to, ShardID, targetShardID, false)
					//	} else {
					//		fmt.Println("这是领导者自身节点，不需要发消息")
					//	}
					//}
				}
			}
		} else {
			fmt.Println("目标分片不存在")
		}
	} else {
		fmt.Println("提议排队中：", proposal.ID)
	}

}
func handleCrossShardProposal(ShardID int, TarGetShardID int, QC QuorumCertificate, proposal Proposal, from string, to string) {
	fmt.Println("handleProposal")
	fmt.Println("ProcessingProposalID", ProcessingProposalID)
	fmt.Println("proposal.ID", proposal.ID)
	if ProcessingProposalID == proposal.ID {
		fmt.Println("处理提议：", proposal.ID)
		//遍历ShardID的关联分片
		fmt.Println("判断关联分片：")
		fmt.Println("len(RelatedSharding[ShardID]) ：", len(RelatedSharding))
		//relatedShardingflag := false
		//crossShardingflag := true

		//if len(RelatedSharding) != 0 {
		//	for _, targetShardID := range RelatedSharding[ShardID] {
		//		if targetShardID == TarGetShardID {
		//			//如果该分片是关联分片，则向该分片领导节点发送消息
		//			fmt.Println("发起分片ID：", ShardID)
		//			fmt.Println("目标分片ID：", targetShardID)
		//			fmt.Println("为关联分片")
		//			relatedShardingflag = true
		//			//广播准备消息给发起分片领导者
		//			SendCrossShardData(knownShardingNodes[ShardID][0], proposal, QC, from, to, ShardID, targetShardID, crossShardingflag, relatedShardingflag)
		//			//广播准备消息给目标分片领导者
		//			SendCrossShardData(knownShardingNodes[targetShardID][0], proposal, QC, from, to, ShardID, targetShardID, crossShardingflag, relatedShardingflag)
		//		} else {
		//			//向两个分片的节点请求验证
		//			fmt.Println("发起分片ID：", ShardID)
		//			fmt.Println("目标分片ID：", targetShardID)
		//			fmt.Println("不是关联分片")
		//		}
		//	}
		//}

	} else {
		fmt.Println("提议排队中：", proposal.ID)
	}
}
func handleSendVoteMsg(request []byte) {
	var buff bytes.Buffer
	var payload VoteMsg
	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	bc := NewBlockchain(NodeIPAddress)
	defer bc.db.Close()
	result := strconv.Itoa(payload.ShardID) + "-" + strconv.Itoa(payload.TarGetShardID)
	//接收到来自其他节点的投票信息，先判断是否是领导者节点
	if NodeIPAddress == knownShardingNodes[payload.ShardID][0] || NodeIPAddress == knownShardingNodes[payload.TarGetShardID][0] || NodeIPAddress == knownShardingNodes[RelatedSharding[result]-1][0] {
		fmt.Println("NodeIP:", NodeIP, "接收到来自", payload.AddrFrom, "的VoteMsg消息")
		fmt.Println("领导者节点，接收投票信息")
		//验证投票信息
		if VerifyByPublicKey(payload.Vote.PublicKey, payload.Message, payload.Vote.R, payload.Vote.S) {
			requiredAgree := (len(knownShardingNodes[belongToInt]) / 2) + 1
			if requiredAgree == 0 {
				requiredAgree = 1
			} else if requiredAgree == 1 {
				requiredAgree = 2
			}
			createOrUpdateVoteCollector(payload.Vote, payload.Message, bc, payload.Proposalvalue, requiredAgree, payload.ShardID, payload.TarGetShardID)
			//if completeproposal[payload.Proposalvalue.ID] == nil {
			//	fmt.Println("验证通过,将投票信息加入投票列表中")
			//	createOrUpdateVoteCollector(payload.Vote, payload.Message, bc, payload.Proposalvalue)
			//} else {
			//	fmt.Println("提议已经完成，不再接收投票信息")
			//}
		} else {
			fmt.Println("验证失败,将投票信息作废")
		}
	} else {
		fmt.Println("非领导者节点，无权限接收投票信息")
	}
}
func handleSendBlockSync(request []byte) {
	var buff bytes.Buffer
	var payload BlockSyncData
	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	fmt.Println("handleSendBlockSync")
	if NodeIPAddress == knownShardingNodes[payload.BelongToInt][0] {
		//接收统计信息
		fmt.Println("NodeIP:", NodeIP, "Received from", payload.AddrFrom, "BlockSync Msg")
		BlockSyncnum++
		fmt.Println("BlockSyncnum:", BlockSyncnum)
		if BlockSyncnum >= len(knownShardingNodes[payload.BelongToInt])-1 {
			fmt.Println("shard:", payload.BelongToInt, "Block synchronization completed")

			if len(proposalpool) > 0 && len(QCpool) > 0 {

				//删除VoteCollector中当前已完成的数据
				delete(voteCollectors, ProcessingProposalID)
				//删除proposalpool中当前已完成的数据
				proposalpool = proposalpool[1:]
				//删除QCpool中当前已完成的数据
				QCpool = QCpool[1:]
				//删除frompool中当前已完成的数据
				frompool = frompool[1:]
				//删除topool中当前已完成的数据
				topool = topool[1:]
				BlockSyncnum = 0
				//fmt.Println("删除proposalpool中当前已完成的数据")
				fmt.Println("Currently, there are：", len(proposalpool), "proposals in the proposal pool")
				fmt.Println("Proposal processing completed:", ProcessingProposalID)
				//如果proposalpool中还有提议，则继续处理
				if len(proposalpool) > 0 {
					fmt.Println("There are still proposals in the proposal pool, continue to process them")
					ProcessingProposalID = proposalpool[0].ID
					from := frompool[0]
					to := topool[0]
					fmt.Println("Next proposal to begin processing:", ProcessingProposalID)
					ShardID := -1
					targetShardID := -1
					//遍历knownShardingNodes[]数组
					fmt.Println("knownShardingNodes:", knownShardingNodes)
					for i, _ := range knownShardingNodes {
						if ShardID >= 0 {
							break
						} else {
							for _, node := range knownShardingNodes[i] {
								fmt.Println("node:", node)
								if node == from {
									ShardID = i
									break
								}
							}
						}
					}
					for i, _ := range knownShardingNodes {
						if targetShardID >= 0 {
							break
						} else {
							for _, node := range knownShardingNodes[i] {
								fmt.Println("node:", node)
								if node == to {
									targetShardID = i
									break
								}
							}
						}
					}

					if ShardID != -1 && targetShardID != -1 {
						handleProposal(ShardID, targetShardID, QCpool[0], proposalpool[0], frompool[0], topool[0])
					}

				} else {
					fmt.Println("There are no proposals in the proposal pool, stop processing-2")
					fmt.Println("Handling ", len(completeproposal), "proposals")
					endTime = time.Now()
					// 计算运行时间
					elapsedTime := endTime.Sub(startTime)
					fmt.Println("Processing proposal runtime：", elapsedTime)
					//清空completeproposal
					completeproposal = make(map[string]*Proposal)
					//清空UsedTxId
					UsedTxId = make(map[string][]byte)
					//bc := NewBlockchain(NodeIPAddress)
					//err := bc.db.View(func(tx *bolt.Tx) error {
					//	b := tx.Bucket([]byte(utxoBucket))
					//	//打印b中的数据
					//	err := b.ForEach(func(k, v []byte) error {
					//		for UsedTxIdKey, UsedTxIdValue := range UsedTxId {
					//			//如果UsedTxId的key和b中的key相同，则打印出来
					//			if hex.EncodeToString(k) == UsedTxIdKey && UsedTxIdValue != nil {
					//				fmt.Println("解锁UTXO -UsedTxId[", UsedTxIdKey, "]=", UsedTxIdValue)
					//				UsedTxId[UsedTxIdKey] = nil
					//			}
					//		}
					//		return nil
					//	})
					//
					//	if err != nil {
					//		log.Panic(err)
					//	}
					//	return nil
					//})
					//bc.db.Close()
					//if err != nil {
					//	log.Panic(err)
					//}

				}
			} else {
				fmt.Println("There are no proposals in the proposal pool, stop processing-1")
			}
		}
	} else {
		fmt.Println("Non leader nodes without permission to receive block synchronization information")
	}

}

func handleSendCrossShardData(request []byte) {
	var buff bytes.Buffer
	var payload CrossShardData
	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	bc := NewBlockchain(NodeIPAddress)
	defer bc.db.Close()
	if payload.RelatedShardingFlag == true {
		fmt.Println("NodeIP:", NodeIP, "接收到来自", payload.AddrFrom, "的CrossShardData消息")
		fmt.Println("领导者节点，接收跨分片信息,这是关联分片交易")

		if VerifyByPublicKey(payload.QC.NodeSignatures[0].PublicKey, payload.QC.Message.Value, payload.QC.NodeSignatures[0].R, payload.QC.NodeSignatures[0].S) {

			var wallet string
			var r1, s1 *big.Int
			//返回签名给领导者节点
			wallet, _, r1, s1, _ = SignByPrivateKey(payload.QC.Message.Value)
			var vote Vote
			vote.Votetype = "agree"
			vote.NodeID = NodeIPAddress
			vote.Addresss = wallet
			vote.R = r1
			vote.S = s1
			vote.PublicKey = publicKey
			vote.Tx = payload.QC.NodeSignatures[0].Tx
			fmt.Println("验证通过,返回签名给领导者节点")
			requiredAgree := 2 //只需要发起分片和目标分片的领导者同意即可
			createOrUpdateVoteCollector(vote, payload.QC.Message.Value, bc, payload.Proposalvalue, requiredAgree, payload.ShardID, payload.TarGetShardID)
			for _, node := range knownShardingNodes[belongToInt] {
				if node != NodeIPAddress {
					node = strings.Replace(node, " ", ":", -1)
					//给分片中除自己外的其他人发送消息
					SendPrepareMsg(node, payload.Proposalvalue, payload.QC, payload.From, payload.To, payload.ShardID, payload.TarGetShardID, payload.CrossShardFlag)
				}
			}
			////如果是发起分片领导者节点，则向目标分片领导者节点发送消息
			//if NodeIPAddress == knownShardingNodes[payload.ShardID][0] {
			//	// 转:
			//	addrIP := strings.Replace(knownShardingNodes[payload.TarGetShardID][0], " ", ":", -1)
			//	//向目标分片领导者节点发送消息
			//	fmt.Println("向目标分片领导者节点发送消息")
			//	sendVoteMsg(addrIP, payload.QC.Message.Value, vote, payload.Proposalvalue, payload.ShardID, payload.TarGetShardID)
			//} else { //如果是目标分片领导者节点，则向发起分片领导者节点发送消息
			//	// 转:
			//	addrIP := strings.Replace(knownShardingNodes[payload.ShardID][0], " ", ":", -1)
			//	//向发起分片领导者节点发送消息
			//	fmt.Println("向发起分片领导者节点发送消息")
			//	sendVoteMsg(addrIP, payload.QC.Message.Value, vote, payload.Proposalvalue, payload.ShardID, payload.TarGetShardID)
			//}
			//fmt.Println("当前节点：", NodeIPAddress, "是分片", belongToInt, "的领导者节点")
		} else {
			fmt.Println("验证不通过")
			var vote Vote
			vote.Votetype = "disagree"
		}
	} else {
		fmt.Println("NodeIP:", NodeIP, "接收到来自", payload.AddrFrom, "的CrossShardData消息")
		fmt.Println("领导者节点，接收跨分片信息,这是非关联分片交易")
		if VerifyByPublicKey(payload.QC.NodeSignatures[0].PublicKey, payload.QC.Message.Value, payload.QC.NodeSignatures[0].R, payload.QC.NodeSignatures[0].S) {

			var wallet string
			var r1, s1 *big.Int
			//返回签名给领导者节点
			wallet, _, r1, s1, _ = SignByPrivateKey(payload.QC.Message.Value)
			var vote Vote
			vote.Votetype = "agree"
			vote.NodeID = NodeIPAddress
			vote.Addresss = wallet
			vote.R = r1
			vote.S = s1
			vote.PublicKey = publicKey
			vote.Tx = payload.QC.NodeSignatures[0].Tx
			fmt.Println("验证通过,返回签名给领导者节点")

			//如果是发起分片领导者节点，则向目标分片领导者节点发送消息
			if NodeIPAddress == knownShardingNodes[payload.ShardID][0] {
				requiredAgree := (len(knownShardingNodes[belongToInt]) / 2) + 1 //只需要发起分片和目标分片的领导者同意即可
				createOrUpdateVoteCollector(vote, payload.QC.Message.Value, bc, payload.Proposalvalue, requiredAgree, payload.ShardID, payload.TarGetShardID)
				// 转:
				addrIP := strings.Replace(knownShardingNodes[payload.TarGetShardID][0], " ", ":", -1)
				//向目标分片领导者节点发送消息
				sendVoteMsg(addrIP, payload.QC.Message.Value, vote, payload.Proposalvalue, payload.ShardID, payload.TarGetShardID)
				//向目标分片的其他节点发送消息
				for _, node := range knownShardingNodes[payload.TarGetShardID] {
					if node != knownShardingNodes[payload.TarGetShardID][0] {
						node = strings.Replace(node, " ", ":", -1)
						//给分片中除领导者外的其他人发送消息
						SendPrepareMsg(node, payload.Proposalvalue, payload.QC, payload.From, payload.To, payload.ShardID, payload.TarGetShardID, payload.CrossShardFlag)
					}
				}
			} else { //如果是目标分片领导者节点，则向发起分片领导者节点发送消息
				requiredAgree := (len(knownShardingNodes[belongToInt]) / 2) + 1 //只需要发起分片和目标分片的领导者同意即可
				createOrUpdateVoteCollector(vote, payload.QC.Message.Value, bc, payload.Proposalvalue, requiredAgree, payload.ShardID, payload.TarGetShardID)
				// 转:
				addrIP := strings.Replace(knownShardingNodes[payload.ShardID][0], " ", ":", -1)
				//向发起分片领导者节点发送消息
				sendVoteMsg(addrIP, payload.QC.Message.Value, vote, payload.Proposalvalue, payload.ShardID, payload.TarGetShardID)
				//向发起分片的其他节点发送消息
				for _, node := range knownShardingNodes[payload.ShardID] {
					if node != knownShardingNodes[payload.ShardID][0] {
						node = strings.Replace(node, " ", ":", -1)
						//给分片中除领导者外的其他人发送消息
						SendPrepareMsg(node, payload.Proposalvalue, payload.QC, payload.From, payload.To, payload.ShardID, payload.TarGetShardID, payload.CrossShardFlag)
					}
				}
			}
			fmt.Println("当前节点：", NodeIPAddress, "是分片", belongToInt, "的领导者节点")
			//向自己分片的其他节点发送消息
			for _, node := range knownShardingNodes[belongToInt] {
				if node != NodeIPAddress {
					node = strings.Replace(node, " ", ":", -1)
					//给分片中除自己外的其他人发送消息
					SendPrepareMsg(node, payload.Proposalvalue, payload.QC, payload.From, payload.To, payload.ShardID, payload.TarGetShardID, payload.CrossShardFlag)
				}
			}
		} else {
			fmt.Println("验证不通过")
			var vote Vote
			vote.Votetype = "disagree"
		}
	}

}

func handleSendBalanceMsg(request []byte) {
	fmt.Println("handleSendBalanceMsg")
	var buff bytes.Buffer
	var payload BalanceData
	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	bc := NewBlockchain(NodeIPAddress)
	defer bc.db.Close()
	fmt.Println("NodeIP:", NodeIP, "接收到来自分片", payload.BelongToInt, "的", payload.AddrFrom, "的BalanceMsg消息")
	fmt.Println("账户:", payload.Address)
	if !ValidateAddress(payload.Address) {
		log.Panic("ERROR: Address is not valid")
	}
	UTXOSet := UTXOSet{bc}
	balance := 0
	pubKeyHash := Base58Decode([]byte(payload.Address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]

	UTXOs := UTXOSet.FindUTXO(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}
	fmt.Println("分片", belongToInt, "的余额为：", balance)
	sendNodeID := strings.Replace(knownShardingNodes[payload.BelongToInt][0], " ", ":", -1)
	sendTotalBalanceMsg(sendNodeID, balance)

}
func handleSendTotalBalanceMsg(request []byte) {
	var buff bytes.Buffer
	var payload TotalBalanceData
	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	fmt.Println("NodeIP:", NodeIP, "接收到来自", payload.AddrFrom, "的TotalBalanceMsg消息")
	fmt.Println("余额:", payload.Balance)
	if payload.AddrFrom != NodeIP {
		totalBalance = totalBalance + payload.Balance
	}
	fmt.Println("totalBalance:", totalBalance)
	countNum++

	var requireCountNum int
	//如果自己是领导者

	requireCountNum = len(knownShardingNodes)

	fmt.Println("countNum", countNum)
	fmt.Println("requireCountNum", requireCountNum)
	if countNum == requireCountNum {

		fmt.Println("totalBalance:", totalBalance)
		fmt.Println("向web后端发送数据")

		data := map[string]interface{}{
			"command":            "StatisticalBalance",
			"balance":            totalBalance,
			"wallet":             "",
			"IP":                 "",
			"port":               "",
			"to":                 "",
			"from":               "",
			"knownShardingNodes": knownShardingNodes,
		}
		// 将数据编码为JSON
		jsonData, err := json.Marshal(data)
		if err != nil {
			log.Fatalf("Error occurred during JSON marshalling: %v", err)
		}
		countNum = 0
		totalBalance = 0
		// 发送POST请求
		resp, err := http.Post("http://192.168.254.129:8088/post", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			log.Fatalf("An Error Occurred: %v", err)
		}
		//defer resp.Body.Close()
		//fmt.Println("Response status:", resp.Status)
		defer resp.Body.Close()
		// 打印响应状态
		log.Println("Response status:", resp.Status)
		//SendBalanceMsg(payload.AddrFrom, address, balance)
	}
}
func handleConnection(conn net.Conn) {

	request, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}

	command := bytesToCommand(request[:commandLength])

	//command := bytesToCommand(request[:commandLength])
	fmt.Printf("Received %s command\n", command)

	switch command {
	case "addr":
		handleAddr(request)
	case "block":
		handleBlock(request)
	case "inv":
		handleInv(request)
	case "getblocks":
		handleGetBlocks(request)
	case "getdata":
		handleGetData(request)
	case "tx":
		handleTx(request)
	case "version":
		handleVersion(request)
	case "sendTestdata":
		handleSendTestdata(request)
	case "sendknownShardingNodes":
		handleSendknownShardingNodes(request)
	case "sendPrepareMsg":
		handleSendPrepareMsg(request)
	case "sendVoteMsg":
		handleSendVoteMsg(request)
	case "sendBlockSync":
		handleSendBlockSync(request)
	case "sendCrossShardData":
		handleSendCrossShardData(request)
	case "sendBalanceMsg":
		handleSendBalanceMsg(request)
	case "sendTotalBalanceMsg":
		handleSendTotalBalanceMsg(request)
	default:
		fmt.Println("Unknown command!")
	}

	conn.Close()
	//如果bc是nil，就不执行bc.db.Close()

	//err = bc.db.Close()
	//if err != nil {
	//	log.Println("Error closing database:", err)
	//}

}

// StartServer 这段代码定义了一个 `StartServer` 函数，用于启动区块链节点的服务器，以监听并处理与其他节点的连接。
//下面是这个函数的功能和步骤解释：
//1. 使用传入的 `nodeID` 构造当前节点的地址 `nodeAddress`，格式为 `localhost:nodeID`。
//2. 设置全局变量 `miningAddress` 为传入的 `minerAddress`，用于指定挖矿奖励的接收地址。
//3. 调用 `net.Listen` 函数在指定的地址上创建监听器，使用指定的网络协议（`protocol` 变量）进行监听。如果出现错误，会触发 Panic。
//4. 在函数执行完毕后，通过 `defer` 关键字关闭监听器，以确保在函数返回之前关闭监听。
//5. 创建一个新的区块链实例 `bc`，使用传入的 `nodeID`。
//6. 如果当前节点的地址不等于已知节点列表中的第一个地址（`knownNodes[0]`），则向已知的某一个节点发送版本信息，以建立连接。
//7. 进入无限循环，等待接受连接请求。当有连接请求到来时，会创建一个新的协程来处理连接，调用 `handleConnection` 函数进行处理，同时传入区块链实例 `bc`。
//总的来说，这个函数的目的是启动一个区块链节点的服务器，用于监听和处理与其他节点的连接，以实现区块链网络的通信和同步。
func StartServer(nodeID, minerAddress string) {
	gob.Register(Proposal{})
	gob.Register(QuorumCertificate{})
	gob.Register(Vote{})
	gob.Register(elliptic.P256())
	substrings := strings.Split(nodeID, " ")

	//replacedString := strings.Replace(nodeID, " ", ":", 0)
	fmt.Println("---------------------------------------")
	fmt.Println("启动节点-nodeID:", nodeID)
	fmt.Println("---------------------------------------")
	//fmt.Println("replacedString String:", replacedString)
	nodeAddress = fmt.Sprintf("%s:%s", substrings[0], substrings[1])
	//输出nodeAddress
	fmt.Println("nodeAddress:", nodeAddress)
	//fmt.Println("knownShardingNodes:", knownShardingNodes)
	miningAddress = minerAddress
	NodeIP = nodeAddress
	NodeIPAddress = strings.Replace(NodeIP, ":", " ", -1)
	//bc := NewBlockchain(nodeID)
	ln, err := net.Listen(protocol, nodeAddress)
	if err != nil {
		log.Panic("ln, err := net.Listen(protocol, nodeAddress)", err)
	}
	defer ln.Close()
	//defer bc.db.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Panic("conn, err := ln.Accept()", err)
		}
		go handleConnection(conn)
	}
}

//这段代码定义了一个 `gobEncode` 函数，用于将数据进行 Gob 编码并返回编码后的字节序列。
//下面是这个函数的功能和步骤解释：
//1. 创建一个新的字节缓冲区 `buff`。
//2. 创建一个 Gob 编码器（`enc`）并将其与字节缓冲区关联。
//3. 使用 Gob 编码器对传入的数据（`data`）进行编码。如果出现错误，会触发 Panic。
//4. 返回字节缓冲区 `buff` 中的字节序列。
//总的来说，这个函数的目的是将给定的数据使用 Gob 编码转换为字节序列，以便在网络通信或持久化存储时使用。
//Gob（Binary Encoder/Decoder）是 Go 语言内置的序列化库，用于将数据编码为字节流以进行传输或存储，并能够在需要时将其解码回原始数据。
func gobEncode(data interface{}) []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

func nodeIsKnown(addr string) bool {
	for _, node := range knownNodes {
		if node == addr {
			return true
		}
	}

	return false
}
