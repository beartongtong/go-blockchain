$BINARY = ".\blockchain_go.exe"
#3000钱包地址
$wallet1 = "1DgDqW4dXk6qtixB9P64yijbUpAHoPYP2o"
$wallet2 = "1AGQPhZeipC4BMWVezmsdNtoZRGdpAqQAA"
$wallet3 = "1LU7FMux1WKhVs2yviye5JL9ExySTk2z9L"
#3001钱包
$wallet4 = "1GZJVbj3xEfa5YYjQMDosdjpVWWJRviCbs"
$wallet5 = "12yUgqMFyiBtVWWW4XMVpJ8Kw2xDZgHieK"
$wallet6 = "17Zqiz4fKGX467R4qBdZ78T57kkmKxDtBF"
#3002矿工钱包地址
$wallet7 = "1MAvKzoqF6jfuphzxaFWbtBFipgfwAiXUE"

# 构建目标
function all {
    build
    test
}

# 构建
function build {
    Write-Host "====> Go build"
    go build -o $BINARY
}
#设置程序端口号
#Set-Item -Path "env:NODE_ID" -Value "3001"
# 测试
function test {
    build
    Write-Host "====>testsend"
    Start-Process -FilePath "blockchain_go.exe"
    #    & $BINARY testsend -sendaddr "192.168.254.130:3000" -data "this is a test data"

}

# 创建钱包
function createWallet {
    Write-Host "====>createWallet"
    & $BINARY createwallet
}
# 创建区块链
function createBlockchain {
    Write-Host "====>createBlockchain"
        & $BINARY createblockchain -address 1DgDqW4dXk6qtixB9P64yijbUpAHoPYP2o
}
# 打印区块链
function printChain {
    Write-Host "====>printChain"
    & $BINARY printchain
}
# 发送交易
function send {
    Write-Host "====>send"
    & $BINARY send -from $wallet4 -to $wallet6 -amount 1
    & $BINARY send -from $wallet5 -to $wallet6 -amount 1
}
# 发送交易并挖矿（仅限于没有矿工节点的时候）
function sendAndMine {
    Write-Host "====>sendAndMine"
    & $BINARY send -from $wallet1 -to $wallet4 -amount 3 -mine
    & $BINARY send -from $wallet1 -to $wallet5 -amount 5 -mine
}
# 查询余额
function getBalance {
    Write-Host "====>getBalance"
    & $BINARY getbalance -address $wallet1
    & $BINARY getbalance -address $wallet2
    & $BINARY getbalance -address $wallet3
    & $BINARY getbalance -address $wallet4
    & $BINARY getbalance -address $wallet5
    & $BINARY getbalance -address $wallet6
    & $BINARY getbalance -address $wallet7
}
#开启节点
function startNode {
    Write-Host "====>startNode"
    & $BINARY startnode
}
#开启矿工节点
function startMinerNode {
    Write-Host "====>startMinerNode"
    & $BINARY startnode -miner $wallet7
}
# 复制数据库
function copyDB {
    Write-Host "====>copyDB"
    Copy-Item -Path .\blockchain_3000.db -Destination .\blockchain_genesis.db
}

# 复制数据库3001
function copyDB3001 {
    Write-Host "====>copyDB3001"
    Copy-Item -Path .\blockchain_genesis.db -Destination .\blockchain_3001.db
}
# 复制数据库3002
function copyDB3002 {
    Write-Host "====>copyDB3002"
    Copy-Item -Path .\blockchain_genesis.db -Destination .\blockchain_3002.db
}# 复制数据库3002
function testscrip {
    Set-Item -Path "env:NODE_ID" -Value "3001"
    Write-Host "Hello from PowerShell"
    Write-Host "====>startNode"
    & .\blockchain_go.exe startnode
}
# 执行构建
#all
Set-Item -Path "env:NODE_ID" -Value "3001"
printChain