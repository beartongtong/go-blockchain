package main

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

type RequestBodyData struct {
	Command            string     `json:"command"`
	IP                 string     `json:"IP"`
	Wallet             string     `json:"wallet"`
	Port               string     `json:"port"`
	To                 string     `json:"to"`
	From               string     `json:"from"`
	KnownShardingNodes [][]string `json:"knownShardingNodes"`
	Balance            int        `json:"balance"`
	Data               string     `json:"data"`
}

var requestFlag = false

var varChangeNotify = make(chan bool) // 创建一个通道，用于通知变量改变
var totalBalance = 0
var countNum = 0
var requestBodyData RequestBodyData

func handleGetRequest(w http.ResponseWriter, r *http.Request) {
	// 获取 URL 中的参数
	queryParams := r.URL.Query()
	paramValue := queryParams.Get("param")
	fmt.Println("Received param:", paramValue)
	// 返回响应
	fmt.Fprintf(w, "Received param: %s", paramValue)
}

func handlePostRequest(w http.ResponseWriter, r *http.Request) {
	//w.Header().Set("Access-Control-Allow-Origin", "http://192.168.254.129") // 替换为你的前端应用的域名或 IP 地址
	w.Header().Set("Access-Control-Allow-Origin", "*") //允许任何源访问，解决跨域问题
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.WriteHeader(http.StatusOK)
	contentType := r.Header.Get("Content-Type")

	containsFormData := strings.Contains(contentType, "multipart/form-data")
	if containsFormData {
		contentType = "multipart/form-data"
	}
	containsUrlencoded := strings.Contains(contentType, "application/x-www-form-urlencoded")
	if containsUrlencoded {
		contentType = "application/x-www-form-urlencoded"
	}
	containsJson := strings.Contains(contentType, "application/json")
	if containsJson {
		contentType = "application/json"
	}

	//if contentType == "" {
	//	http.Error(w, "Content-Type header is not present", http.StatusBadRequest)
	//	return
	//}
	fmt.Println("Content-Type:", contentType)

	switch contentType {
	case "application/x-www-form-urlencoded":
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		formData := r.PostForm
		requestBodyData.Command = formData.Get("command")
		requestBodyData.IP = formData.Get("IP")
		requestBodyData.Wallet = formData.Get("wallet")
		//fmt.Fprintf(w, "Received param (application/x-www-form-urlencoded): %v", formData)
	case "multipart/form-data":
		err := r.ParseMultipartForm(10 << 20) // 10 MB limit
		if err != nil {
			http.Error(w, "Failed to parse form-data", http.StatusBadRequest)
			return
		}

		//file, _, err := r.FormFile("file") // 假设你有一个名为file的文件输入字段"
		//if err != nil {
		//	http.Error(w, "Failed to get file", http.StatusBadRequest)
		//	//return
		//} else {
		//	fmt.Println("获取file成功")
		//	//	输出file内容
		//	buf := bytes.NewBuffer(nil)
		//	if _, err := io.Copy(buf, file); err != nil {
		//		log.Fatal(err)
		//	}
		//	fmt.Println("输出file内容", buf.String())
		//
		//	defer file.Close()
		//	fmt.Fprintf(w, "输出file内容", buf.String())
		//}

		//获取post所有数据
		formData := r.PostForm

		requestBodyData.Command = formData.Get("command")
		requestBodyData.IP = formData.Get("IP")
		requestBodyData.Wallet = formData.Get("wallet")
		//fmt.Fprintf(w, "Received param (multipart/form-data): %v", formData)

	case "application/json":
		if r.Body == nil {
			http.Error(w, "Empty request body", http.StatusBadRequest)
			return
		}
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&requestBodyData)
		if err != nil {
			http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
			return
		}
		fmt.Println("decoder:", decoder)
		fmt.Println("requestBodyData:", requestBodyData)
		fmt.Println("requestBodyData.Command:", requestBodyData.Command)
		fmt.Println("requestBodyData.IP:", requestBodyData.IP)
		fmt.Println("requestBodyData.Wallet:", requestBodyData.Wallet)
		fmt.Println("requestBodyData.Port:", requestBodyData.Port)
		fmt.Println("requestBodyData.To:", requestBodyData.To)
		fmt.Println("requestBodyData.From:", requestBodyData.From)
		fmt.Println("requestBodyData.KnownShardingNodes:", requestBodyData.KnownShardingNodes)
		fmt.Println("requestBodyData.Balance:", requestBodyData.Balance)
		fmt.Println("requestBodyData.Data:", requestBodyData.Data)
	default:
		http.Error(w, "Unsupported Content-Type", http.StatusBadRequest)
		return
	}
	NodeIP = requestBodyData.IP + ":" + requestBodyData.Port
	NodeIPAddress = requestBodyData.IP + " " + requestBodyData.Port
	fmt.Println("Command:", requestBodyData.Command)
	fmt.Println("IP:", requestBodyData.IP)
	fmt.Println("Wallet:", requestBodyData.Wallet)
	fmt.Println("Port:", requestBodyData.Port)
	// 构建单条 JSON 数据
	//singleData := map[string]interface{}{
	//	"wallet": "walletaddres",
	//}

	// 将单条 JSON 数据编码为 JSON 格式的字节数组
	//jsonData, err := json.Marshal(singleData)
	//if err != nil {
	//	http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
	//	return
	//}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")

	// 返回 JSON 数据
	//w.WriteHeader(http.StatusOK)
	//w.Write(jsonData)

	// 使用逗号作为分隔符拆分字符串
	substrings := strings.Split(requestBodyData.Command, " ")
	//nodeID := strings.Split(IP, " ")
	fmt.Println(substrings)

	//for _, s := range substrings {
	//	fmt.Println(s)
	//}
	switch substrings[0] {
	case "getbalance":
		fmt.Println("getbalance")
		address := substrings[2]
		//IP拼接Port
		nodeID := requestBodyData.IP + " " + requestBodyData.Port
		if !ValidateAddress(address) {
			log.Panic("ERROR: Address is not valid")
		}

		bc := NewBlockchain(nodeID)
		UTXOSet := UTXOSet{bc}

		//defer bc.db.Close()

		balance := 0
		pubKeyHash := Base58Decode([]byte(address))
		pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]

		UTXOs := UTXOSet.FindUTXO(pubKeyHash)

		for _, out := range UTXOs {
			balance += out.Value
		}

		fmt.Printf("Balance of '%s': %d\n", address, balance)
		fmt.Fprintf(w, "Balance of '%s': %d\n", address, balance)
		bc.db.Close()
	case "createblockchain":
		fmt.Println("createblockchain")
		fmt.Println(substrings[1])
		address := substrings[1]
		nodeID := requestBodyData.IP + " " + requestBodyData.Port
		if !ValidateAddress(address) {
			log.Panic("ERROR: Address is not valid")
		}
		bc := CreateBlockchain(address, nodeID)
		defer bc.db.Close()

		UTXOSet := UTXOSet{bc}
		UTXOSet.Reindex()

		fmt.Println("Done!")
	case "createsubblockchain":
		fmt.Println("createsubblockchain")
		nodeID := requestBodyData.IP + " " + requestBodyData.Port

		dbFile := fmt.Sprintf(dbFile, nodeID)
		if !dbExists(dbFile) {
			fmt.Println("Blockchain not exists.")
			genesisdbFile := "blockchain_genesis.db"
			if !dbExists(genesisdbFile) {
				fmt.Println("blockchain_genesis.db not exists.")
				os.Exit(1)
			}

			// 源文件和目标文件名
			sourceFile := genesisdbFile
			targetFile := dbFile

			// 打开源文件
			source, err := os.Open(sourceFile)
			if err != nil {
				fmt.Println("Error opening source file:", err)
				return
			}
			defer source.Close()

			// 创建目标文件
			target, err := os.Create(targetFile)
			if err != nil {
				fmt.Println("Error creating target file:", err)
				return
			}
			defer target.Close()

			// 复制文件内容
			_, err = io.Copy(target, source)
			if err != nil {
				fmt.Println("Error copying file:", err)
				return
			}

			fmt.Println("File copied and renamed successfully.")
			fmt.Fprintf(w, "File copied and renamed successfully.")
		}
	case "createwallet":
		fmt.Println("createwallet")
		var walletstr string
		//如果substrings[1]为空，就不传入参数
		if substrings[1] == "" {
			walletstr = requestBodyData.IP + " " + requestBodyData.Port
		} else {
			walletstr = substrings[1]
		}
		wallets, _ := NewWallets(walletstr)
		address := wallets.CreateWallet()
		wallets.SaveToFile(walletstr)

		fmt.Printf("Your new address: %s\n", address)
		fmt.Println("PowerShell script executed successfully!")
		singleData := map[string]interface{}{
			"walletaddress": address,
		}
		// 将单条 JSON 数据编码为 JSON 格式的字节数组
		jsonData, err := json.Marshal(singleData)
		if err != nil {
			http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
	case "listaddresses":
		fmt.Println("listaddresses")
		wallets, err := NewWallets(requestBodyData.IP + " " + requestBodyData.Port)
		if err != nil {
			log.Panic(err)
		}
		addresses := wallets.GetAddresses()

		//for _, address := range addresses {
		//	fmt.Println(address)
		//}

		singleData := map[string]interface{}{
			"walletaddress": addresses,
		}
		//打印输出singleData
		fmt.Println("singleData", singleData)
		// 将单条 JSON 数据编码为 JSON 格式的字节数组
		jsonData, err := json.Marshal(singleData)
		if err != nil {
			http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)

	case "printchain":
		fmt.Println("printchain")
		bc := NewBlockchain(requestBodyData.IP + " " + requestBodyData.Port)
		defer bc.db.Close()

		bci := bc.Iterator()
		fmt.Println("bci:", bci)
		var blocks []map[string]interface{}

		for {
			block := bci.Next()

			blockInfo := map[string]interface{}{
				"Hash":      hex.EncodeToString(block.Hash),
				"Height":    block.Height,
				"PrevBlock": hex.EncodeToString(block.PrevBlockHash),
				"Data":      block.Transactions,
				"Timestamp": block.Timestamp,
				"Nonce":     block.Nonce,

				// ... 提取其他你需要的块信息
			}
			fmt.Println("blockInfo:", blockInfo)
			fmt.Println("blockInfo[Hash]:", blockInfo["Hash"])
			fmt.Println("blockInfo[Data]:", blockInfo["Data"])
			fmt.Println("blockInfo[Height]:", blockInfo["Height"])
			fmt.Println("blockInfo[PrevBlock]:", blockInfo["PrevBlock"])
			fmt.Println("blockInfo[Nonce]:", blockInfo["Nonce"])
			blocks = append(blocks, blockInfo)

			if len(block.PrevBlockHash) == 0 {
				break
			}
		}

		responseData := map[string]interface{}{
			"Blocks": blocks,
		}

		jsonData, err := json.Marshal(responseData)
		if err != nil {
			http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
	case "reindexutxo":
		fmt.Println("reindexutxo")
	case "send":
		fmt.Println("send")
		from := substrings[2]
		to := substrings[4]
		amount, err := strconv.Atoi(substrings[6])
		//nodeID := requestBodyData.IP + " " + requestBodyData.Port
		nodeID := requestBodyData.From
		mineNow := false
		if substrings[7] == "mine" {
			mineNow = true
			fmt.Println("mine now：", mineNow)
		}
		if !ValidateAddress(from) {
			log.Panic("ERROR: Sender address is not valid")
		}
		if !ValidateAddress(to) {
			log.Panic("ERROR: Recipient address is not valid")
		}
		bc := NewBlockchain(nodeID)
		UTXOSet := UTXOSet{bc}
		fmt.Println("UTXOSet", UTXOSet)
		defer bc.db.Close()

		wallets, err := NewWallets(nodeID)
		if err != nil {
			log.Panic(err)
		}
		fmt.Println("wallets", wallets)
		wallet := wallets.GetWallet(from)
		fmt.Println("wallet", wallet)
		tx := NewUTXOTransaction(&wallet, to, amount, &UTXOSet)
		fmt.Println("tx", tx)
		if mineNow {
			cbTx := NewCoinbaseTX(from, "")
			txs := []*Transaction{cbTx, tx}

			newBlock := bc.MineBlock(txs)
			UTXOSet.Update(newBlock)
		} else {
			sendTx(knownShardingNodes[0][0], tx)
		}

		fmt.Println("Success!")
	case "startnode":
		fmt.Println("startnode")
		//0是领导节点  1是普通节点 2是挖矿节点
		Isleader := 0
		//新增的节点属于哪个分片
		belongTo = substrings[3]
		belongToInt, _ = strconv.Atoi(belongTo)

		fmt.Println("startnode-IP", requestBodyData.IP+" "+requestBodyData.Port)
		nodeID := requestBodyData.IP + " " + requestBodyData.Port
		binary := "./blockchain_go.exe"
		//binary := "./go_build_blockchain_go.exe"
		script := fmt.Sprintf(`$BINARY = "%s"
						function startNode {
							 Write-Host "====>startNode"
						  & $BINARY startnode
						}
					Set-Item -Path "env:NODE_ID" -Value "%s"
					startNode`, binary, nodeID)
		if substrings[1] == "-miner" {
			Isleader = 2

			script = fmt.Sprintf(`$BINARY = "%s"
						function startNode {
							 Write-Host "====>startNode"
						  & $BINARY startnode -miner "%s"
						}
					Set-Item -Path "env:NODE_ID" -Value "%s"
					startNode`, binary, substrings[2], nodeID)
		}
		if substrings[1] == "-leader" {
			fmt.Println("startnode leader")
			Isleader = 0
			//int转string
			belongTo = strconv.Itoa(len(knownShardingNodes))
			element := []string{}

			knownShardingNodes = append(knownShardingNodes, element)

			//knownShardingNodes[0] = append(knownShardingNodes[0], IP)
			//knownShardingNodes[0] = append(knownShardingNodes[0], IP)
			//knownShardingNodes[0] = append(knownShardingNodes[0], IP)

			knownShardingNodes[len(knownShardingNodes)-1] = append(knownShardingNodes[len(knownShardingNodes)-1], nodeID)
			//遍历knownShardingNodes

			//response := fmt.Sprintf("分片%d添加成功", len(knownShardingNodes))
			//fmt.Fprintf(w, response)
		}
		if substrings[1] == "-normal" {
			Isleader = 1
			//判断knownShardingNodes[0]是否已经存在
			//如果存在，就不添加
			if len(knownShardingNodes) == 0 {
				http.Error(w, "knownShardingNodes is empty 分片领导者节点为空", http.StatusInternalServerError)
				//fmt.Fprintf(w, "knownShardingNodes is empty 分片领导者节点为空")
				return
			} else {
				//string转int
				belongToInt, _ := strconv.Atoi(belongTo)
				knownShardingNodes[belongToInt] = append(knownShardingNodes[belongToInt], nodeID)
				fmt.Println(knownShardingNodes[belongToInt][0])
			}

		}
		fmt.Println(knownShardingNodes)

		ps1FilePath := "myscript.ps1"
		// 将指定内容写入.ps1文件
		err := ioutil.WriteFile(ps1FilePath, []byte(script), 0644)
		if err != nil {
			fmt.Println("Error writing to .ps1 file:", err)
			return
		}
		fmt.Printf("Script content written to %s\n", ps1FilePath)
		scriptPath := "myscript.ps1"
		// 使用 cmd.exe 打开 PowerShell 窗口并执行脚本文件
		cmd := exec.Command("cmd", "/C", "start", "powershell", "-NoExit", "-File", scriptPath)

		err = cmd.Start()
		SendknownShardingNodes(requestBodyData.IP+":"+requestBodyData.Port, belongToInt)
		if err != nil {
			fmt.Println("Error starting PowerShell:", err)
			return
		}

		if substrings[1] == "-miner" {
			//fmt.Fprintf(w, "矿工节点启动成功")
		}
		if substrings[1] == "-leader" {
			//fmt.Fprintf(w, " 领导节点启动成功")
		}
		if substrings[1] == "-normal" {
			//fmt.Fprintf(w, "普通节点启动成功")
		}
		//以json格式返回knownShardingNodes和Isleader参数
		singleData := map[string]interface{}{
			"knownShardingNodes": knownShardingNodes,
			"Isleader":           Isleader,
			"belongTo":           belongTo,
		}
		// 将单条 JSON 数据编码为 JSON 格式的字节数组
		jsonData, err := json.Marshal(singleData)
		if err != nil {
			http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
		fmt.Println("PowerShell script executed successfully!")

	case "testsend":
		fmt.Println("testsend")
		// 空格替换成:号
		//nodeIP := strings.Replace(IP, " ", ":", -1)
		//SendknownShardingNodes(nodeIP)
		//fmt.Println(knownNodes)
		//knownNodes = append(knownNodes, "192.168.254.129:3000")
		//knownNodes = append(knownNodes, "192.168.254.129:3001")
		//element1 := []string{}
		//element2 := []string{"X", "Y", "Z"}
		//knownShardingNodes = append(knownShardingNodes, element1)
		//
		//knownShardingNodes = append(knownShardingNodes, element2)
		//knownShardingNodes[0] = append(knownShardingNodes[0], "d")
		//knownShardingNodes[0] = append(knownShardingNodes[0], "1111")
		//fmt.Println(knownNodes)
		fmt.Println(knownShardingNodes)
		//fmt.Println(knownShardingNodes[0])
		//fmt.Println(knownShardingNodes[0][0])
		//fmt.Println(knownShardingNodes[0][1])
		//fmt.Println(knownShardingNodes[1])
		//fmt.Println(knownShardingNodes[1][0])
		//fmt.Println(knownShardingNodes[1][1])
		//输出knownNodes
	case "verifySignature":
		fmt.Println("verifySignature")
		address := substrings[2]
		message := substrings[4]
		nodeID := requestBodyData.IP + " " + requestBodyData.Port
		fmt.Println(address)
		fmt.Println(message)
		//message string 转[]byte
		messageByte := []byte(message)
		wallets, err := NewWallets(nodeID)
		if err != nil {
			log.Panic(err)
		}
		//创建一个*big.Int类型变量
		var r1, s1 *big.Int
		address, messageByte, r1, s1, _ = wallets.Sign(address, messageByte)
		if wallets.Verify(address, message, r1, s1) {
			fmt.Println("签名验证成功")
		} else {
			fmt.Println("签名验证失败")
		}
	case "createLeaf":
		fmt.Println("createLeaf")
		fmt.Println("Command:", requestBodyData.Command)
		fmt.Println("IP:", requestBodyData.IP)
		fmt.Println("Wallet:", requestBodyData.Wallet)
		//定义一个Node变量
		node := Node{
			ID:             requestBodyData.IP + " " + requestBodyData.Port,
			Proposals:      []*Proposal{},
			LastProposalID: "",
			Mutex:          sync.Mutex{},
		}
		proposal := CreateLeaf(&node, requestBodyData.Command, requestBodyData.Wallet)

		//commandstring转[]byte
		commandByte := []byte(requestBodyData.Command)
		wallets, err := NewWallets(requestBodyData.IP + " " + requestBodyData.Port)
		if err != nil {
			log.Panic(err)
		}
		var r1, s1 *big.Int
		requestBodyData.Wallet, commandByte, r1, s1, _ = wallets.Sign(requestBodyData.Wallet, commandByte)
		if wallets.Verify(requestBodyData.Wallet, requestBodyData.Command, r1, s1) {
			fmt.Println("签名验证成功")
		} else {
			fmt.Println("签名验证失败")
		}
		//定义一个Vote变量
		var vote Vote
		vote.NodeID = requestBodyData.IP + " " + requestBodyData.Port
		vote.Addresss = requestBodyData.Wallet
		vote.R = r1
		vote.S = s1
		CreateQC(vote, *proposal)
	case "createProposal":
		fmt.Println("createProposal")
		fmt.Println("Command:", requestBodyData.Command)
		oldSubstring := " "
		newSubstring := ":"
		sendNodeID := strings.Replace(requestBodyData.From, oldSubstring, newSubstring, -1)
		fmt.Println("---knownShardingNodes===:", knownShardingNodes)
		sendTestdata(sendNodeID, requestBodyData.Command, requestBodyData.From, requestBodyData.To)
		//if substrings[1] == "getbalance" {
		//	//给所有分片领导者发消息
		//	//遍历knownShardingNodes
		//	sendNodeID = strings.Replace(requestBodyData.From, oldSubstring, newSubstring, -1)
		//	sendTestdata(sendNodeID, requestBodyData.Command, requestBodyData.From, requestBodyData.To)
		//	//for _, knownShardingNode := range knownShardingNodes {
		//	//	//遍历knownShardingNode
		//	//	fmt.Println("向领导者：", knownShardingNode[0], "发送消息")
		//	//	sendNodeID = strings.Replace(knownShardingNode[0], oldSubstring, newSubstring, -1)
		//	//	//给分片领导者发消息
		//	//
		//	//}
		//	// 然后等待变量改变的通知
		//	select {
		//	case <-varChangeNotify:
		//		fmt.Println("????????4:")
		//		// 一旦接收到通知，发送响应
		//		fmt.Fprintf(w, "totalBalance of '%s': %d\n", substrings[3], totalBalance)
		//		totalBalance = 0
		//		countNum = 0
		//		requestFlag = false
		//	case <-time.After(5 * time.Second):
		//		// 或者超时后返回，客户端将再次请求
		//		//fmt.Fprintf(w, "Timeout, please poll again")
		//		fmt.Println("totalBalance", totalBalance)
		//		fmt.Fprintf(w, "totalBalance of '%s': %d\n", substrings[3], totalBalance)
		//	}
		//} else {
		//	sendTestdata(sendNodeID, requestBodyData.Command, requestBodyData.From, requestBodyData.To)
		//}

	case "UTXOSet":
		bc := NewBlockchain(requestBodyData.IP + " " + requestBodyData.Port)
		UTXOSet := UTXOSet{bc}
		db := UTXOSet.Blockchain.db
		err := db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(utxoBucket))
			//打印b中的数据
			err := b.ForEach(func(k, v []byte) error {

				value := b.Get(k)
				//fmt.Printf("Key: %s, Value: %s\n", hex.EncodeToString(k), hex.EncodeToString(value))
				fmt.Printf("Key: %s, Value: %s\n", hex.EncodeToString(k))
				//fmt.Printf(" %s\n", hex.EncodeToString(k))
				var outputs TXOutputs

				if len(value) == 0 {
					log.Panic("DeserializeOutputs: Empty data")
				}

				dec := gob.NewDecoder(bytes.NewReader(value))
				err := dec.Decode(&outputs)
				if err != nil {
					log.Panic("err := dec.Decode(&outputs)", err)
				}
				fmt.Println("outputs", outputs)
				//遍历打印outputs
				for _, output := range outputs.Outputs {
					fmt.Println("output.Value", output.Value)
					fmt.Println("output.PubKeyHash", output.PubKeyHash)
				}
				return nil
			})
			return nil
			if err != nil {
				log.Panic(err)
			}
			return nil
		})
		if err != nil {
			log.Panic(err)
		}
		bc.db.Close()
	case "vin.Txid":
		bc := NewBlockchain(requestBodyData.IP + " " + requestBodyData.Port)
		UTXOSet := UTXOSet{bc}
		db := UTXOSet.Blockchain.db
		//[]byte类型变量转string

		//创建一个[]byte类型变量 并赋值33653163353430653364653366363931303361316563393862633563613865316163306363643037616137386263313162343233373631393533626266346638，然后转为string
		//hexString := "33653163353430653364653366363931303361316563393862633563613865316163306363643037616137386263313162343233373631393533626266346638"

		// 将 []byte 转换为字符串
		//strData := string(byteData)

		//fmt.Println("strData", strData)
		strData := substrings[1]
		fmt.Println("substrings", substrings)
		fmt.Println("substrings[1]", substrings[1])
		fmt.Println("strData", strData)
		//将十六进制字符串解码为 []byte
		byteData, err := hex.DecodeString(substrings[1])
		if err != nil {
			fmt.Println("解码失败:", err)
			return
		}
		//TXid := []byte(strData)
		err = db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(utxoBucket))
			outsBytes := b.Get(byteData)
			//fmt.Printf("UTXOSet.Update - vin.Txid: %s, outsBytes: %v\n", hex.EncodeToString(Base58Decode(byteData)), hex.EncodeToString(Base58Decode(outsBytes)))
			fmt.Printf("UTXOSet.Update - vin.Txid: %s, outsBytes: %v\n", hex.EncodeToString(byteData), hex.EncodeToString(outsBytes))
			var outputs TXOutputs

			if len(outsBytes) == 0 {
				log.Panic("DeserializeOutputs: Empty data")
			}

			dec := gob.NewDecoder(bytes.NewReader(outsBytes))
			err := dec.Decode(&outputs)
			if err != nil {
				log.Panic("err := dec.Decode(&outputs)", err)
			}
			fmt.Println("outputs", outputs)
			//遍历打印outputs
			for _, output := range outputs.Outputs {
				fmt.Println("output.Value", output.Value)
				fmt.Println("output.PubKeyHash", output.PubKeyHash)
			}
			return nil
		})
		if err != nil {
			log.Panic(err)
		}
		//关闭数据库
		bc.db.Close()
	case "initKnownShardingNodes":
		fmt.Println("initKnownShardingNodes")
		knownShardingNodes = requestBodyData.KnownShardingNodes
		fmt.Println("knownShardingNodes:", knownShardingNodes)

		singleData := map[string]interface{}{
			"knownShardingNodes": knownShardingNodes,
			"belongTo":           belongTo,
		}
		// 将单条 JSON 数据编码为 JSON 格式的字节数组
		jsonData, err := json.Marshal(singleData)
		if err != nil {
			http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
		//knownShardingNodes = append(knownShardingNodes, "

	case "DistributeRewards":
		fmt.Println("DistributeRewards")
		bc := NewBlockchain(requestBodyData.IP + " " + requestBodyData.Port)
		fmt.Println("1")
		address := substrings[2]
		amount, err := strconv.Atoi(substrings[3])
		//nodeID := requestBodyData.IP + " " + requestBodyData.Port
		var newSourceBlock *Block
		if err != nil {
			log.Panic(err)
		}
		if !ValidateAddress(address) {
			fmt.Println("ERROR: Address is not valid")
		} else {
			cbtx := NewCoinbaseTX2(address, requestBodyData.Command, amount)

			fmt.Println("2")
			sourcetxs := []*Transaction{cbtx}

			sourceUTXOSet := UTXOSet{bc} //发起分片

			newSourceBlock = bc.commitTransaction(sourcetxs, requestBodyData.Data)
			fmt.Println("3")
			sourceUTXOSet.Update(newSourceBlock)
			fmt.Println("4")
			fmt.Printf("----Added block %x\n", newSourceBlock.Hash)
			fmt.Printf("----Added block %x\n", string(newSourceBlock.Hash))
			fmt.Printf("----Added block %x\n", hex.EncodeToString(newSourceBlock.Hash))
		}
		fmt.Println("区块链上链成功!")
		//返回json数据
		singleData := map[string]interface{}{
			"Hash": hex.EncodeToString(newSourceBlock.Hash),
		}
		// 将单条 JSON 数据编码为 JSON 格式的字节数组
		jsonData, err := json.Marshal(singleData)
		if err != nil {
			http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		//打印json数据

		w.Write(jsonData)
		//关闭数据库
		fmt.Println("5")
		bc.db.Close()
		//var shardID int
		////遍历knownShardingNodes
		//for index, knownShardingNode := range knownShardingNodes {
		//	//遍历knownShardingNode
		//	for _, node := range knownShardingNode {
		//		//查找nodeID
		//		if node == nodeID {
		//			shardID = index
		//		}
		//	}
		//}
		//
		//
		//for _, node := range knownShardingNodes[shardID] {
		//	if node == knownShardingNodes[shardID][0] {
		//		//fmt.Println("我是领导者")
		//	} else {
		//
		//		node = strings.Replace(node, " ", ":", -1)
		//		//fmt.Println("给子节点更新区块：", node)
		//		BlockSyncnum = 0
		//		//广播区块
		//		sendVersion(node, bc, shardID)
		//	}
		//}

	case "StatisticalBalance":
		fmt.Println("-=-=-=StatisticalBalance-=-=-=-")
		totalBalance = requestBodyData.Balance
		fmt.Println("totalBalance", totalBalance)
		requestFlag = true
		varChangeNotify <- requestFlag // 通知变量改变
	case "getBlock":
		fmt.Println("getBlock")
		bc := NewBlockchain(requestBodyData.IP + " " + requestBodyData.Port)
		defer bc.db.Close()

		bci := bc.Iterator()
		fmt.Println("bci:", bci)
		var blockInfo map[string]interface{}

		for {
			block := bci.Next()

			if len(block.PrevBlockHash) == 0 || hex.EncodeToString(block.Hash) == substrings[1] {
				blockInfo = map[string]interface{}{
					"Hash":      hex.EncodeToString(block.Hash),
					"Height":    block.Height,
					"PrevBlock": hex.EncodeToString(block.PrevBlockHash),
					//"Data":      hex.EncodeToString(block.Data),
					"Data":      string(block.Data),
					"Timestamp": block.Timestamp,
					"Nonce":     block.Nonce,
				}

				break
			}
		}

		fmt.Println("blockInfo[Hash]:", blockInfo["Hash"])
		fmt.Println("blockInfo[Data]:", blockInfo["Data"])
		fmt.Println("blockInfo[Height]:", blockInfo["Height"])
		fmt.Println("blockInfo[PrevBlock]:", blockInfo["PrevBlock"])
		fmt.Println("blockInfo[Nonce]:", blockInfo["Nonce"])
		responseData := map[string]interface{}{
			"Hash":      blockInfo["Hash"],
			"Height":    blockInfo["Height"],
			"PrevBlock": blockInfo["PrevBlock"],
			"Data":      blockInfo["Data"],
			"Timestamp": blockInfo["Timestamp"],
			"Nonce":     blockInfo["Nonce"],
		}

		jsonData, err := json.Marshal(responseData)
		if err != nil {
			http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
	default:
		fmt.Println("default")
		fmt.Fprintf(w, "Received param (default): %s,", substrings[0])
		fmt.Fprintf(w, "Usage:")
		fmt.Fprintf(w, "createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
		fmt.Fprintf(w, "createwallet - Generates a new key-pair and saves it into the wallet file")
		fmt.Fprintf(w, "getbalance -address ADDRESS - Get balance of ADDRESS")
		fmt.Fprintf(w, "listaddresses - Lists all addresses from the wallet file")
		fmt.Fprintf(w, "printchain - Print all the blocks of the blockchain")
		fmt.Fprintf(w, "reindexutxo - Rebuilds the UTXO set")
		fmt.Fprintf(w, "send -from FROM -to TO -amount AMOUNT -mine - Send AMOUNT of coins from FROM address to TO. Mine on the same node, when -mine is set.")
		fmt.Fprintf(w, "startnode -miner ADDRESS - Start a node with ID specified in NODE_ID env. var. -miner enables mining")
		fmt.Fprintf(w, "testsend -data ADDRESS - Send test data to ADDRESS")

	}
}
