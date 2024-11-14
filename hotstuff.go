package main

import (
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"
)

//Vote 表示HotStuff中的投票包含投票者的地址和签名数据
type Vote struct {
	Votetype  string
	NodeID    string
	Addresss  string
	S         *big.Int
	R         *big.Int
	PublicKey ecdsa.PublicKey
	Tx        *Transaction
}

// VoteCollector 用于收集投票信息
type VoteCollector struct {
	mu                   sync.Mutex
	votes                map[string]Vote
	totalVotes           int
	requiredAgrees       int
	leaderAgress         int
	requiredleaderAgress int
	once                 sync.Once
}

// Proposal 表示HotStuff中的提案
type Proposal struct {
	ID       string //提案者的节点IP
	Value    string
	Proposer string // 提案者的节点钱包地址
}

// Node 表示HotStuff中的节点
type Node struct {
	ID             string
	Proposals      []*Proposal
	LastProposalID string // 最后一个提案的ID
	Mutex          sync.Mutex
}

type QuorumCertificate struct {
	ViewNumber     int
	Type           string
	NodeSignatures map[int]Vote // 节点签名
	Message        Proposal     // 消息或提案，具体类型取决于您的应用
}

//Msg 表示HotStuff中的消息
type Msg struct {
	msgtype           int
	node              string
	quorumCertificate *QuorumCertificate
}

//Replica 是表示 HotStuff 协议中的复制节点（replica）的数据结构。

type Replica struct {
	ID         string
	curView    int
	leaderID   string
	NextViewCh chan int
}

var voteCollectors = make(map[string]*VoteCollector)
var IndexOfCbtx = 0

//创建提案
func CreateLeaf(node *Node, command string, wallet string) *Proposal {
	//获取当前时间
	//currentTime := time.Now().Format(time.RFC3339Nano)
	currentTime := time.Now().UTC().Format("2006-01-02T15:04:05.999999999")
	//fmt.Println("currentTime：", currentTime)
	// 创建提案
	proposal := &Proposal{
		ID:       belongTo + "-" + node.ID + fmt.Sprintf("-%s ", currentTime),
		Value:    command,
		Proposer: wallet,
	}
	// 将提案添加到节点的提案列表中
	node.Proposals = append(node.Proposals, proposal)
	// 更新节点的最后一个提案ID
	node.LastProposalID = proposal.ID
	//fmt.Println("proposal.ID：", proposal.ID)
	// 使用逗号作为分隔符拆分字符串
	lastProposerNum := strings.Split(proposal.ID, "-")
	fmt.Println("lastProposerNum：", lastProposerNum[1])
	//fmt.Println("proposal.Value：", proposal.Value)
	//fmt.Println("proposal.Proposer：", proposal.Proposer)
	return proposal
}

// CreateQC 创建证明
func CreateQC(votes Vote, proposal Proposal) *QuorumCertificate {

	// 创建证明
	qc := &QuorumCertificate{
		ViewNumber: 0,
		Type:       "PrePrepare",

		Message: proposal,
	}
	//往NodeSignatures中添加初始节点签名 和对应的address 以及签名数据
	qc.NodeSignatures = make(map[int]Vote)
	//fmt.Println("len：", len(qc.NodeSignatures))
	qc.NodeSignatures[0] = Vote{"agree", votes.NodeID, votes.Addresss, votes.S, votes.R, votes.PublicKey, votes.Tx}
	//fmt.Println("CreateQC：", qc.NodeSignatures[0])
	//fmt.Println("len：", len(qc.NodeSignatures))
	//fmt.Println("创建证明成功")
	return qc
}

// 判断是否是领导者
func amILeader(leaderID string, curView int) bool {
	//遍历knownShardingNodes[]
	for i := 0; i < len(knownShardingNodes); i++ {
		if knownShardingNodes[i][0] == leaderID {
			return true
		}
	}
	return false
}

// NewVoteCollector 创建一个新的投票收集器
func NewVoteCollector(requiredAgrees int, totalVotes int, leaderAgress int, requiredleaderAgress int) *VoteCollector {
	return &VoteCollector{
		votes:                make(map[string]Vote),
		requiredAgrees:       requiredAgrees,
		totalVotes:           totalVotes,
		leaderAgress:         leaderAgress,
		requiredleaderAgress: requiredleaderAgress,
	}
}

// AddVote 添加一个投票到收集器中
func (vc *VoteCollector) AddVote(vote Vote, command string, bc *Blockchain, proposal Proposal, shardID int, targetShardID int) {
	vc.mu.Lock()
	defer vc.mu.Unlock()

	// 检查是否已经有该节点的投票
	if existingVote, ok := vc.votes[vote.NodeID]; ok {
		// 如果已经有投票，可以在这里处理冲突
		fmt.Printf("Node %s already voted: %v\n", vote.NodeID, existingVote)
		return
	}

	// 添加新的投票
	vc.votes[vote.NodeID] = vote
	vc.totalVotes++
	//判断是不是领导者投票
	if amILeader(vote.NodeID, vc.totalVotes) {
		fmt.Println("领导者:", vote.NodeID, "投票")
		vc.leaderAgress++
	}
	fmt.Printf("Received vote from Node %s: \n", vote.NodeID)
	fmt.Printf("收到来自节点 %s: 的消息\n", vote.NodeID)
	fmt.Println("当前投票总数：", vc.totalVotes)
	fmt.Println("所需同意的数量：", vc.requiredAgrees)
	fmt.Println("当前领导者同意的数量：", vc.leaderAgress)
	fmt.Println("所需领导者同意的数量：", vc.requiredleaderAgress)

	//超出的票数就不接收了，防止同一个提议被多次执行
	if vc.totalVotes == vc.requiredAgrees && vc.leaderAgress == vc.requiredleaderAgress {
		fmt.Println("Consensus reached! Perform necessary actions.")
		if targetShardID == shardID {
			fmt.Println("分片内交易")
			fmt.Println("更新本分片", shardID, "数据")

			completeproposal[proposal.ID] = &proposal
			// 在这里执行达成共识后的操作
			substrings := strings.Split(command, " ")
			fmt.Println("command", command)
			switch substrings[1] {
			case "send":
				//达成共识后数据上链
				from := substrings[3]
				to := substrings[5]
				if !ValidateAddress(from) {
					log.Panic("ERROR: Sender address is not valid")
				}
				if !ValidateAddress(to) {
					log.Panic("ERROR: Recipient address is not valid")
				}
				fmt.Println("if !ValidateAddress(to) ")
				//UTXOSet := UTXOSet{shardIDbc}

				cbTx := NewCoinbaseTX(from, "这是系统奖励，开发新区块10元")

				IndexOfCbtx++
				//fmt.Println("-=-=-=-=-=-==-=-=IndexOfCbtx-=-=-=-=-=-==-=-=", IndexOfCbtx)
				if vote.Tx == nil {
					fmt.Println("ERROR: Received vote with nil transaction")
					return
				}

				sourcetxs := []*Transaction{cbTx, vote.Tx}

				var newSourceBlock *Block

				sourceUTXOSet := UTXOSet{bc} //发起分片

				newSourceBlock = bc.commitTransaction(sourcetxs, command)

				fmt.Println("----UTXOSet.Update(newSourceBlock)")
				sourceUTXOSet.Update(newSourceBlock)

				fmt.Println("区块链上链成功!")
				for _, node := range knownShardingNodes[belongToInt] {
					if node == knownShardingNodes[belongToInt][0] {
						//fmt.Println("我是领导者")
					} else {

						node = strings.Replace(node, " ", ":", -1)
						//fmt.Println("给子节点更新区块：", node)
						BlockSyncnum = 0
						//广播区块
						sendVersion(node, bc, shardID)
					}
				}
			}
		} else {
			fmt.Println("跨分片交易")
			//判断是不是关联分片
			relatedShardingflag := false
			//for _, relatedShardID := range RelatedSharding[shardID] {
			//	if relatedShardID == targetShardID {
			//		relatedShardingflag = true
			//		break
			//	}
			//}
			//拼接字符串
			result := strconv.Itoa(shardID) + "-" + strconv.Itoa(targetShardID)
			if RelatedSharding[result] > 0 {
				relatedShardingflag = true
			}
			fmt.Println("-----result-------", result)
			fmt.Println("-----RelatedSharding[result]-------", RelatedSharding[result])
			fmt.Println("-----RelatedSharding[0-1]]-------", RelatedSharding["0-1"])
			if relatedShardingflag {
				fmt.Println("关联分片交易")

				var sourceShardIDbc *Blockchain
				var targetShardIDbc *Blockchain
				fmt.Println("dbFile", dbFile)
				fmt.Println("创建分片", shardID, "数据库连接", knownShardingNodes[shardID][0])
				sourceShardIDbc = NewBlockchain(knownShardingNodes[shardID][0]) //发起分片的数据库
				fmt.Println("创建分片", targetShardID, "数据库连接", knownShardingNodes[targetShardID][0])
				targetShardIDbc = NewBlockchain(knownShardingNodes[targetShardID][0]) //发起分片的数据库
				completeproposal[proposal.ID] = &proposal
				// 在这里执行达成共识后的操作
				substrings := strings.Split(command, " ")
				fmt.Println("command", command)
				switch substrings[1] {
				case "send":

					//达成共识后数据上链
					from := substrings[3]
					to := substrings[5]

					amount, _ := strconv.Atoi(substrings[7])
					if !ValidateAddress(from) {
						log.Panic("ERROR: Sender address is not valid")
					}
					if !ValidateAddress(to) {
						log.Panic("ERROR: Recipient address is not valid")
					}
					fmt.Println("if !ValidateAddress(to) ")
					//UTXOSet := UTXOSet{shardIDbc}

					cbTx := NewCoinbaseTX(from, "")
					tarGetcbTx := NewCoinbaseTX2(to, "", amount)
					IndexOfCbtx++
					//fmt.Println("-=-=-=-=-=-==-=-=IndexOfCbtx-=-=-=-=-=-==-=-=", IndexOfCbtx)
					if vote.Tx == nil {
						fmt.Println("ERROR: Received vote with nil transaction")
						return
					}

					sourcetxs := []*Transaction{cbTx, vote.Tx}
					targGettxs := []*Transaction{tarGetcbTx}
					//打印交易
					//遍历cbTx.Vin

					var newSourceBlock *Block
					var newTargGetBlock *Block

					sourceUTXOSet := UTXOSet{sourceShardIDbc}  //发起分片
					targGetUTXOSet := UTXOSet{targetShardIDbc} //目标分片
					fmt.Println("----newSourceBlock = bc.commitTransaction(sourcetxs)")
					newSourceBlock = sourceShardIDbc.commitTransaction(sourcetxs, command)
					fmt.Printf("----Added block %x\n", newSourceBlock.Hash)

					fmt.Println("----newTargGetBlock = shardIDbc.commitTransaction(targGettxs)")
					newTargGetBlock = targetShardIDbc.commitTransaction(targGettxs, command)
					fmt.Printf("----Added block %x\n", newTargGetBlock.Hash)

					fmt.Println("----UTXOSet.Update(newSourceBlock)")
					sourceUTXOSet.Update(newSourceBlock)

					fmt.Println("----UTXOSet.Update(newTargGetBlock)")
					targGetUTXOSet.Update(newTargGetBlock)

					fmt.Println("区块链上链成功!")

					fmt.Println("更新发起分片数据")
					//node := strings.Replace(knownShardingNodes[shardID][0], " ", ":", -1)
					//sendVersion(node, sourceShardIDbc, shardID)
					for _, node := range knownShardingNodes[shardID] {
						node = strings.Replace(node, " ", ":", -1)
						fmt.Println("给子节点更新区块：", node)
						BlockSyncnum = 0
						//广播区块
						sendVersion(node, sourceShardIDbc, shardID)
					}
					fmt.Println("更新目标分片数据")
					for _, node := range knownShardingNodes[targetShardID] {
						node = strings.Replace(node, " ", ":", -1)
						fmt.Println("给子节点更新区块：", node)
						BlockSyncnum = 0
						//广播区块
						sendVersion(node, targetShardIDbc, targetShardID)
					}
					defer sourceShardIDbc.db.Close()
					defer targetShardIDbc.db.Close()
				}

			} else {
				fmt.Println("非关联分片交易")
				if belongToInt == shardID {
					fmt.Println("更新本分片", shardID, "数据")
					//var targetShardIDbc *Blockchain
					//targetShardIDbc = NewBlockchain(knownShardingNodes[targetShardID][0]) //发起分片的数据库
					completeproposal[proposal.ID] = &proposal
					// 在这里执行达成共识后的操作
					substrings := strings.Split(command, " ")
					fmt.Println("command", command)
					switch substrings[1] {
					case "send":

						//达成共识后数据上链
						from := substrings[3]
						to := substrings[5]

						//amount, _ := strconv.Atoi(substrings[7])
						if !ValidateAddress(from) {
							log.Panic("ERROR: Sender address is not valid")
						}
						if !ValidateAddress(to) {
							log.Panic("ERROR: Recipient address is not valid")
						}

						//UTXOSet := UTXOSet{shardIDbc}

						cbTx := NewCoinbaseTX(from, "")
						//tarGetcbTx := NewCoinbaseTX2(to, "", amount)
						IndexOfCbtx++
						//fmt.Println("-=-=-=-=-=-==-=-=IndexOfCbtx-=-=-=-=-=-==-=-=", IndexOfCbtx)
						if vote.Tx == nil {
							fmt.Println("ERROR: Received vote with nil transaction")
							return
						}

						sourcetxs := []*Transaction{cbTx, vote.Tx}
						//targGettxs := []*Transaction{tarGetcbTx}
						//打印交易
						//遍历cbTx.Vin

						var newSourceBlock *Block
						//var newTargGetBlock *Block

						sourceUTXOSet := UTXOSet{bc} //发起分片
						//targGetUTXOSet := UTXOSet{targetShardIDbc} //目标分片
						fmt.Println("----newSourceBlock = bc.commitTransaction(sourcetxs)")
						newSourceBlock = bc.commitTransaction(sourcetxs, command)
						//fmt.Println("----newTargGetBlock = shardIDbc.commitTransaction(targGettxs)")
						//newTargGetBlock = targetShardIDbc.commitTransaction(targGettxs)
						fmt.Println("----UTXOSet.Update(newSourceBlock)")
						sourceUTXOSet.Update(newSourceBlock)
						//fmt.Println("----UTXOSet.Update(newTargGetBlock)")
						//targGetUTXOSet.Update(newTargGetBlock)

						fmt.Println("区块链上链成功!")
						//targetShardIDbc.db.Close()
						for _, node := range knownShardingNodes[belongToInt] {
							if node == knownShardingNodes[belongToInt][0] {
								//fmt.Println("我是领导者")
							} else {

								node = strings.Replace(node, " ", ":", -1)
								//fmt.Println("给子节点更新区块：", node)
								BlockSyncnum = 0
								//广播区块
								sendVersion(node, bc, belongToInt)
							}
						}
					}
				} else { //如果targetShardID == belongToInt 代表这是目标分片，需要更新的发起分片的数据
					fmt.Println("更新本分片", targetShardID, "数据")
					//fmt.Println("更新关联分片", shardID, "数据")
					//var shardIDbc *Blockchain
					//shardIDbc = NewBlockchain(knownShardingNodes[shardID][0]) //发起分片的数据库
					completeproposal[proposal.ID] = &proposal
					// 在这里执行达成共识后的操作
					substrings := strings.Split(command, " ")
					fmt.Println("command", command)
					switch substrings[1] {
					case "send":

						//达成共识后数据上链
						from := substrings[3]
						to := substrings[5]

						amount, _ := strconv.Atoi(substrings[7])
						if !ValidateAddress(from) {
							log.Panic("ERROR: Sender address is not valid")
						}
						if !ValidateAddress(to) {
							log.Panic("ERROR: Recipient address is not valid")
						}

						//UTXOSet := UTXOSet{shardIDbc}

						//cbTx := NewCoinbaseTX(from, "")
						tarGetcbTx := NewCoinbaseTX2(to, command, amount)
						IndexOfCbtx++
						//fmt.Println("-=-=-=-=-=-==-=-=IndexOfCbtx-=-=-=-=-=-==-=-=", IndexOfCbtx)
						if vote.Tx == nil {
							fmt.Println("ERROR: Received vote with nil transaction")
							return
						}

						//sourcetxs := []*Transaction{cbTx, vote.Tx}
						targGettxs := []*Transaction{tarGetcbTx}
						//打印交易
						//遍历cbTx.Vin

						//var newSourceBlock *Block
						var newTargGetBlock *Block

						//sourceUTXOSet := UTXOSet{shardIDbc} //发起分片
						targGetUTXOSet := UTXOSet{bc} //目标分片
						//fmt.Println("----newSourceBlock = bc.commitTransaction(sourcetxs)")
						//newSourceBlock = shardIDbc.commitTransaction(sourcetxs)
						fmt.Println("----newTargGetBlock = shardIDbc.commitTransaction(targGettxs)")
						newTargGetBlock = bc.commitTransaction(targGettxs, command)
						//fmt.Println("----UTXOSet.Update(newSourceBlock)")
						//sourceUTXOSet.Update(newSourceBlock)
						fmt.Println("----UTXOSet.Update(newTargGetBlock)")
						targGetUTXOSet.Update(newTargGetBlock)

						fmt.Println("区块链上链成功!")
						fmt.Println("belongToInt", belongToInt)
						//shardIDbc.db.Close()
						for _, node := range knownShardingNodes[belongToInt] {
							if node == knownShardingNodes[belongToInt][0] {
								//fmt.Println("我是领导者")
							} else {

								node = strings.Replace(node, " ", ":", -1)
								//fmt.Println("给子节点更新区块：", node)
								BlockSyncnum = 0
								//广播区块
								sendVersion(node, bc, belongToInt)
							}
						}

					}
				}
			}
		}

		////添加完成的提议信息到completeproposal
		//completeproposal[proposal.ID] = &proposal
		//// 在这里执行达成共识后的操作
		//substrings := strings.Split(command, " ")
		//fmt.Println("command", command)
		//switch substrings[1] {
		//case "send":
		//
		//	//达成共识后数据上链
		//	from := substrings[3]
		//	to := substrings[5]
		//
		//	if !ValidateAddress(from) {
		//		log.Panic("ERROR: Sender address is not valid")
		//	}
		//	if !ValidateAddress(to) {
		//		log.Panic("ERROR: Recipient address is not valid")
		//	}
		//	fmt.Println("if !ValidateAddress(to) ")
		//	UTXOSet := UTXOSet{bc}
		//
		//	cbTx := NewCoinbaseTX(from, "")
		//	IndexOfCbtx++
		//	//fmt.Println("-=-=-=-=-=-==-=-=IndexOfCbtx-=-=-=-=-=-==-=-=", IndexOfCbtx)
		//	if vote.Tx == nil {
		//		fmt.Println("ERROR: Received vote with nil transaction")
		//		return
		//	}
		//	txs := []*Transaction{cbTx, vote.Tx}
		//
		//	fmt.Println("txs := []*Transaction{cbTx, vote.Tx}")
		//	var newBlock *Block
		//	newBlock = bc.commitTransaction(txs)
		//
		//	UTXOSet.Update(newBlock)
		//	fmt.Println("区块链上链成功!")
		//	//fmt.Println("belongToInt", belongToInt)
		//	for _, node := range knownShardingNodes[belongToInt] {
		//		if node == knownShardingNodes[belongToInt][0] {
		//			//fmt.Println("我是领导者")
		//		} else {
		//
		//			node = strings.Replace(node, " ", ":", -1)
		//			//fmt.Println("给子节点更新区块：", node)
		//			BlockSyncnum = 0
		//			//广播区块
		//			sendVersion(node, bc)
		//		}
		//	}
		//
		//}

	}
}

// createOrUpdateVoteCollector 根据节点消息动态创建或更新 VoteCollector
func createOrUpdateVoteCollector(vote Vote, command string, bc *Blockchain, proposal Proposal, requiredAgree int, shardID int, targetShardID int) {
	fmt.Println("-===================-----------------------=========================")
	fmt.Println("createOrUpdateVoteCollector-proposal.ID", proposal.ID)
	fmt.Println("-===================-----------------------=========================")
	// 检查映射中是否已经有对应的 VoteCollector
	if collector, ok := voteCollectors[proposal.ID]; ok {
		// 如果已经存在，直接添加投票
		fmt.Println("")
		collector.AddVote(vote, command, bc, proposal, shardID, targetShardID)
	} else {
		// 如果不存在，创建新的 VoteCollector
		fmt.Println("创建新的 VoteCollector")
		if shardID == targetShardID {
			fmt.Println("非跨分片交易")
			newCollector := NewVoteCollector(requiredAgree, 0, 0, 1) // 你可以根据实际情况调整 requiredAgrees
			voteCollectors[proposal.ID] = newCollector

			fmt.Println("进行投票1")
			newCollector.AddVote(vote, command, bc, proposal, shardID, targetShardID)
		} else {
			fmt.Println("跨分片交易")
			result := strconv.Itoa(shardID) + "-" + strconv.Itoa(targetShardID)
			var newCollector *VoteCollector
			if RelatedSharding[result] > 0 {
				requiredAgree = (len(knownShardingNodes[RelatedSharding[result]-1]) / 2) + 1
				newCollector = NewVoteCollector(requiredAgree, 0, 0, 1) // 你可以根据实际情况调整 requiredAgrees
			} else {
				requiredAgree = (((len(knownShardingNodes[shardID]) + len(knownShardingNodes[targetShardID])) * 8) / 10)
				newCollector = NewVoteCollector(requiredAgree, 0, 0, 2) // 你可以根据实际情况调整 requiredAgrees
			}

			voteCollectors[proposal.ID] = newCollector

			fmt.Println("进行投票2")
			newCollector.AddVote(vote, command, bc, proposal, shardID, targetShardID)
		}

	}
}

//准备阶段
func preparePhase(leaderID string, curView int, node *Node, command string, wallet string, tx *Transaction, from string, to string) {

	proposal := CreateLeaf(node, command, wallet)
	commandByte := []byte(command)
	//fmt.Println("创建当前提案成功 command:", command)
	//fmt.Println("node", node)
	wallets, err := NewWallets(node.ID)
	if err != nil {
		log.Panic(err)
	}
	//fmt.Println("读取钱包成功", wallets)
	var r1, s1 *big.Int
	wallet, commandByte, r1, s1, _ = wallets.Sign(wallet, commandByte)
	//fmt.Println("签名成功")
	if wallets.Verify(wallet, command, r1, s1) {
		fmt.Println("签名验证成功")
		//定义一个Vote变量
		var vote Vote
		vote.Votetype = "agree"
		vote.NodeID = node.ID //节点ID，无：号，带空格
		vote.Addresss = wallet
		vote.R = r1
		vote.S = s1
		vote.PublicKey = publicKey
		vote.Tx = tx
		QC := CreateQC(vote, *proposal)
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
				//广播准备消息
				SendPrepareMsg(leaderID, *proposal, *QC, from, to, belongToInt, targetShardID, false)
			} else {
				fmt.Println("跨分片交易")
				fmt.Println("发起分片ID：", belongToInt)
				fmt.Println("目标分片ID：", targetShardID)

				//广播准备消息
				SendPrepareMsg(leaderID, *proposal, *QC, from, to, belongToInt, targetShardID, true)

			}
		} else {
			fmt.Println("目标分片不存在")
		}
	} else {
		fmt.Println("签名验证失败")
	}

	//addr := "192.168.254.129 3000"

	//if amILeader(leaderID, curView) {
	//	fmt.Println("我是领导者")
	//	//作为领导者，等待新视图消息
	//	//newViewMessages := waitForNewViewMessages(curView)
	//	//选择高QC
	//	//highQC := selectHighQC(newViewMessages)
	//	//创建当前提案
	//}
	//else {
	//	fmt.Println("我是副本")
	//	//作为副本，等待准备消息
	//	prepareMsg := waitForPrepareMessage(curView)
	//
	//	// 验证消息并投票
	//	if isSafeNode(prepareMsg.Node, prepareMsg.Justify) {
	//		sendVoteMsg(prepareMsg.Node, curView)
	//	}
	//}
}

// 预提交阶段
//func preCommitPhase(leaderID string, curView int) {
//	if amILeader(leaderID, curView) {
//		// 作为领导者，等待投票
//		votes := waitForVotes(curView)
//
//		// 创建预准备QC
//		prepareQC := CreateQC(votes)
//
//		// 广播预提交消息
//		broadcastPreCommitMsg(prepareQC)
//	} else {
//		// 作为副本，等待预提交消息
//		preCommitMsg := waitForPreCommitMessage(curView)
//
//		// 验证消息并投票
//		sendVoteMsg(preCommitMsg.Justify.Node, curView)
//	}
//}
//
//// 提交阶段
//func commitPhase(leaderID string, curView int) {
//	if amILeader(leaderID, curView) {
//		// 作为领导者，等待投票
//		votes := waitForVotes(curView)
//
//		// 创建预提交QC
//		preCommitQC := CreateQC(votes)
//
//		// 广播提交消息
//		broadcastCommitMsg(preCommitQC)
//	} else {
//		// 作为副本，等待提交消息
//		commitMsg := waitForCommitMessage(curView)
//
//		// 验证消息并投票
//		sendVoteMsg(commitMsg.Justify.Node, curView)
//	}
//}

//
//// 决策阶段
//func decidePhase(leaderID string, curView int) {
//	if amILeader(leaderID, curView) {
//		// 作为领导者，等待投票
//		votes := waitForVotes(curView)
//
//		// 创建提交QC
//		commitQC := CreateQC(votes)
//
//		// 广播决策消息
//		broadcastDecideMsg(commitQC)
//	} else {
//		// 作为副本，等待决策消息
//		decideMsg := waitForDecideMessage(curView)
//
//		// 执行新命令并响应客户端
//		executeAndRespond(decideMsg.Justify.Node)
//	}
//}

// 下一个视图中断
//func nextViewInterrupt(curView int) {
//	// 在任何阶段的“等待”期间，如果调用了nextView(curView)，则跳转到这一行
//	// 向下一个领导者发送新视图消息
//	sendNewViewMsg(curView + 1)
//}
