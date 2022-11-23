package myclient

import (
	"BlockchainInGo/addresses"
	"BlockchainInGo/blockchain"
	"BlockchainInGo/constcoe"
	"BlockchainInGo/proto"
	"BlockchainInGo/test"
	"BlockchainInGo/transaction"
	"BlockchainInGo/utils"
	"BlockchainInGo/wallet"
	"bytes"
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io/ioutil"
	"log"
	"os"
	"time"
)

var addrsAlive = []string{}

// SendAA 用于向myserver同步addrsAlive
func SendAA() []string {
	return addrsAlive
}

var rdb *redis.Client

func initRedis() (err error) {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379", // 指定
		Password: "",
		DB:       6, // redis一共16个库，指定其中一个库即可
	})
	_, err = rdb.Ping().Result()
	return
}

func RedisInit() {
	err := initRedis()
	if err != nil {
		fmt.Printf("connect redis failed! err : %v\n", err)
		return
	}
	//fmt.Println("redis连接成功！")
}

func BroadcastAlive(client proto.BlockchainServiceClient) bool {
	_, err := client.Alive(context.Background(), &proto.AliveRequest{})
	if err != nil {
		log.Printf("Could not broadcast Alive: %v\n", err)
		return false
	}
	return true
}
func QuickTest() {
	addrsAlive = append(addrsAlive, "0.0.0.0:8001")
	addresses.SaveNewAddress("0.0.0.0:8001")
}

// TestAlive 系统启动时要先检查哪些结点在线
func TestAlive() {
	for i := 0; i < len(addrsAlive); i++ {
		conn, err := grpc.Dial(addrsAlive[i], grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("Failed to connect: %v\n", err)
		}
		client := proto.NewBlockchainServiceClient(conn)
		if !BroadcastAlive(client) {
			addresses.DelPort(addrsAlive[i])
		}
		defer conn.Close()
	}
}

// DeleteAlive 当发现有结点无法连接时要从在线结点列表中去除它的地址
func DeleteAlive(addr string) {
	var tmp []string
	for _, add := range addrsAlive {
		if add != addr {
			tmp = append(tmp, add)
		}
	}
	addrsAlive = tmp
}

func BB(client proto.BlockchainServiceClient, block blockchain.Block, address string) {
	if !BroadCastBlock(client, block) {
		DeleteAlive(address)
		addresses.DelPort(address)
		return
	}
	if !BroadCastTransactions(client, block) {
		DeleteAlive(address)
		addresses.DelPort(address)
		return
	}
	if !BroadCastInputs(client, block) {
		DeleteAlive(address)
		addresses.DelPort(address)
		return
	}
	if !BroadCastOutputs(client, block) {
		DeleteAlive(address)
		addresses.DelPort(address)
		return
	}
	if !BroadcastEnd(client) {
		DeleteAlive(address)
		addresses.DelPort(address)
		return
	}
}

func BroadcastCreator(client proto.BlockchainServiceClient, name string) bool {
	_, err := client.CreateBy(context.Background(), &proto.CreateByRequest{
		Name: name,
	})
	if err != nil {
		log.Printf("Could not broadcast Creator: %v\n", err)
		return false
	}
	return true
}

// BroadCastBlock 广播区块
func BroadCastBlock(client proto.BlockchainServiceClient, block blockchain.Block) bool {
	//log.Println("BroadCastBlock was invoked")
	_, err := client.Block(context.Background(), &proto.BlockRequest{
		Index:      block.Index,
		Timestamp:  block.Timestamp,
		Hash:       block.Hash,
		PrevHash:   block.PrevHash,
		Difficulty: block.Difficulty,
		Target:     block.Target,
		Nonce:      block.Nonce,
		Creator:    block.Creator,
	})
	if err != nil {
		log.Printf("Could not broadcast block: %v\n", err)
		return false
	}
	//log.Println("Block broadcast:", r.BR)
	return true
}

// BroadCastTransactions 广播交易
func BroadCastTransactions(client proto.BlockchainServiceClient, block blockchain.Block) bool {
	//log.Println("BroadCastTransactions was invoked")
	stream, err := client.Transactions(context.Background())
	if err != nil {
		log.Printf("Error while calling LongGreet %v\n", err)
		return false
	}
	txs := []*proto.TransactionsRequest{}
	for i := 0; i < len(block.Transactions); i++ {
		ptx := proto.TransactionsRequest{
			BelongHash: block.Hash,
			ID:         block.Transactions[i].ID,
			Fee:        block.Transactions[i].Fee,
			Type:       block.Transactions[i].Type,
			From:       block.Transactions[i].From,
			To:         block.Transactions[i].TO,
			Amount:     block.Transactions[i].Amount,
		}
		txs = append(txs, &ptx)
	}
	for _, tx := range txs {
		//log.Printf("Sending transaction:%x\n", tx.ID)
		stream.Send(tx)
	}
	_, err = stream.CloseAndRecv()
	if err != nil {
		log.Printf("Error while receiving response from LongGreet :%v\n", err)
		return false
	}
	//log.Println("Transactions result:", res)
	return true
}

// BroadCastInputs 广播交易输入
func BroadCastInputs(client proto.BlockchainServiceClient, block blockchain.Block) bool {
	//log.Println("BroadCastInputs was invoked")
	stream, err := client.Inputs(context.Background())
	if err != nil {
		log.Printf("Error while calling LongGreet %v\n", err)
		return false
	}
	inputs := []*proto.InputsRequest{}
	for _, tx := range block.Transactions {
		for i, input := range tx.Inputs {
			pInput := proto.InputsRequest{
				BelongId: tx.ID,
				Index:    int64(i),
				TxID:     input.TxID,
				OutIdx:   input.OutIdx,
				PubKey:   input.PubKey,
				Sig:      input.Sig,
			}
			inputs = append(inputs, &pInput)
		}
	}
	for _, inp := range inputs {
		//log.Printf("Sending input:%x\n,from:%x", inp.Index, inp.BelongId)
		stream.Send(inp)
	}
	_, err = stream.CloseAndRecv()
	if err != nil {
		log.Printf("Error while receiving response from LongGreet :%v\n", err)
		return false
	}
	//log.Println("Transactions result:", res)
	return true
}

// BroadCastOutputs 广播交易输出
func BroadCastOutputs(client proto.BlockchainServiceClient, block blockchain.Block) bool {
	//log.Println("BroadCastOutputs was invoked")
	stream, err := client.Outputs(context.Background())
	if err != nil {
		log.Printf("Error while calling LongGreet %v\n", err)
		return false
	}
	outputs := []*proto.OutputsRequest{}
	for _, tx := range block.Transactions {
		for i, output := range tx.TxOutput {
			pOutput := proto.OutputsRequest{
				BelongId:   tx.ID,
				Index:      int64(i),
				Value:      int64(output.Value),
				HashPubKey: output.HashPubKey,
			}
			outputs = append(outputs, &pOutput)
		}
	}
	for _, outp := range outputs {
		//log.Printf("Sending output:%x\n,from:%x", outp.Index, outp.BelongId)
		stream.Send(outp)
	}
	_, err = stream.CloseAndRecv()
	if err != nil {
		log.Printf("Error while receiving response from LongGreet :%v\n", err)
		return false
	}
	//log.Println("Transactions result:", res)
	return true
}

// BroadcastEnd 广播传输完毕信号，对方收到后可以根据收到的区块、交易、交易输入输出重新组合出区块，验证后加入区块链
func BroadcastEnd(client proto.BlockchainServiceClient) bool {
	//log.Println("BroadCastEnd was invoked")
	r, err := client.End(context.Background(), &proto.EndRequest{
		EndFlag: true,
	})
	if err != nil {
		log.Printf("Could not broadcast block: %v\n", err)
		return false
	}
	log.Println(r.ER)
	return true
}

// BroadcastGenBlock 广播创世区块（创世区块比较特殊，要单独广播，对方收到后根据创世区块生成区块链）
func BroadcastGenBlock(block blockchain.Block, i int) {
	//addrs := addresses.ReadAllAddress()
	conn, err := grpc.Dial(addrsAlive[i], grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Failed to connect: %v\n", err)
		return
	}
	defer conn.Close()
	client := proto.NewBlockchainServiceClient(conn)
	// 先判断结点在线再广播
	if !BroadcastCreator(client, constcoe.Refname) {
	}
	BB(client, block, addrsAlive[i])

}

// Init 清空所有钱包和区块链信息，生成miner钱包，生成创世区块，广播给所有结点
func Init() {
	os.RemoveAll("tmp/blocks")
	os.RemoveAll("D:/tmp/wallets")
	os.RemoveAll("D:/tmp/ref_list")
	os.Mkdir("tmp/blocks", os.ModePerm)
	os.Mkdir("D:/tmp/wallets", os.ModePerm)
	os.Mkdir("D:/tmp/ref_list", os.ModePerm)
}

func BlockchainInit() {
	block := test.CreateGenBlockRefName(constcoe.Refname)
	// 广播创世区块
	//addrs := addresses.ReadAllAddress()
	for i := 0; i < len(addrsAlive); i++ {
		BroadcastGenBlock(*block, i)
	}
	//test.ShowWallets()
}

func BeginMining() {
	address := test.WalletAddress(constcoe.Refname)
	for {
		txp := blockchain.CreateTransactionPool()
		if len(txp.PubTx) == 0 {
			//println("Mining...")
			time.Sleep(3 * time.Second)
			continue
		}
		block := test.Mine(address, utils.Address2PubHash([]byte(address)))
		// 这prehash传过来必出问题，print都报错，在test.Mine里print一点事都没有。而且hash无论在哪里print也都一点事没有，看了俩小时也不知道为什么只能改成用读写文件传递了
		prehash, readError := ioutil.ReadFile("the weirdest fucking bug.txt")
		if readError != nil {
			panic("sth wrong")
		}
		block.PrevHash = prehash
		//addrs := addresses.ReadAllAddress()
		for i := 0; i < len(addrsAlive); i++ {
			conn, err := grpc.Dial(addrsAlive[i], grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				log.Printf("Failed to connect: %v\n", err)
				return
			}
			defer conn.Close()
			// 创建新客户端
			client := proto.NewBlockchainServiceClient(conn)
			// 先判断结点在线再广播
			BB(client, block, addrsAlive[i])
			//fmt.Printf("hash in block:%x\n", block.Hash)
			//fmt.Printf("Client_hash:%x\n", block.CalculateHash())
			//test.ShowWallets()
		}
	}
}

func LongMiningTest(round int, Tx func()) {
	address := test.WalletAddress(constcoe.Refname)
	for i := 0; i < round; i++ {
		Tx()
		txp := blockchain.CreateTransactionPool()
		if len(txp.PubTx) == 0 {
			//println("Mining...")
			time.Sleep(3 * time.Second)
			continue
		}
		block := test.Mine(address, utils.Address2PubHash([]byte(address)))
		// 这prehash传过来必出问题，print都报错，在test.Mine里print一点事都没有。而且hash无论在哪里print也都一点事没有，看了俩小时也不知道为什么只能改成用读写文件传递了
		prehash, readError := ioutil.ReadFile("the weirdest fucking bug.txt")
		if readError != nil {
			panic("sth wrong")
		}
		block.PrevHash = prehash
		//addrs := addresses.ReadAllAddress()
		for i := 0; i < len(addrsAlive); i++ {
			conn, err := grpc.Dial(addrsAlive[i], grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				log.Printf("Failed to connect: %v\n", err)
			}
			defer conn.Close()
			// 创建新客户端
			client := proto.NewBlockchainServiceClient(conn)
			// 先判断结点在线再广播
			BB(client, block, addrsAlive[i])
			//fmt.Printf("hash in block:%x\n", block.Hash)
			//fmt.Printf("Client_hash:%x\n", block.CalculateHash())
			//test.ShowWallets(true)
		}
	}
}

func GetBlock(client proto.BlockchainServiceClient, height int64) {
	res, err := client.GetBlock(context.Background(), &proto.GetBlockRequest{
		Height:  height,
		Address: constcoe.Address,
	})
	if err != nil {
		log.Printf("Could not broadcast wallet: %v\n", err)
	}
	println(res.WR)
}

func GetBlocks() {
	conn, err := grpc.Dial(addrsAlive[0], grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Failed to connect: %v\n", err)
	}
	defer conn.Close()
	client := proto.NewBlockchainServiceClient(conn)

	var CHeight int64
	if !blockchain.BlockchainExist() {
		CHeight = 0
	} else {
		CHeight = blockchain.ReadLB().Index
	}
	println("CHeight:", CHeight)

	GetBlock(client, CHeight)
}

func InitAll() {
	addresses.PortInit()
	Init()
	wallet.CreateWallet("A")
	wallet.CreateWallet("B")
	wallet.CreateWallet("C")
	addrsAlive = addresses.AlivePort()
	BlockchainInit()
}

func NewTransaction(client proto.BlockchainServiceClient, tx transaction.Transaction) {
	res, err := client.NewTransaction(context.Background(), &proto.TransactionsRequest{
		BelongHash: nil,
		ID:         tx.ID,
		Fee:        tx.Fee,
		Type:       tx.Type,
		From:       tx.From,
		To:         tx.TO,
		Amount:     tx.Amount,
	})
	if err != nil {
		log.Fatalf("Could not broadcast wallet: %v\n", err)
	}
	println(res.TR)
}

func NewOutputs(client proto.BlockchainServiceClient, tx transaction.Transaction) {
	//log.Println("BroadCastOutputs was invoked")
	stream, err := client.NewOutPuts(context.Background())
	if err != nil {
		log.Fatalf("Error while calling LongGreet %v\n", err)
	}
	outputs := []*proto.OutputsRequest{}

	for i, output := range tx.TxOutput {
		pOutput := proto.OutputsRequest{
			BelongId:   tx.ID,
			Index:      int64(i),
			Value:      int64(output.Value),
			HashPubKey: output.HashPubKey,
		}
		outputs = append(outputs, &pOutput)
	}

	for _, outp := range outputs {
		//log.Printf("Sending output:%x\n,from:%x", outp.Index, outp.BelongId)
		stream.Send(outp)
	}
	_, err = stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error while receiving response from LongGreet :%v\n", err)
	}
	//log.Println("Transactions result:", res)
}

func NewInputs(client proto.BlockchainServiceClient, tx transaction.Transaction) {
	//log.Println("BroadCastInputs was invoked")
	stream, err := client.NewInPuts(context.Background())
	if err != nil {
		log.Fatalf("Error while calling LongGreet %v\n", err)
	}
	inputs := []*proto.InputsRequest{}
	for i, input := range tx.Inputs {
		pInput := proto.InputsRequest{
			BelongId: tx.ID,
			Index:    int64(i),
			TxID:     input.TxID,
			OutIdx:   input.OutIdx,
			PubKey:   input.PubKey,
			Sig:      input.Sig,
		}
		inputs = append(inputs, &pInput)
	}

	for _, inp := range inputs {
		//log.Printf("Sending input:%x\n,from:%x", inp.Index, inp.BelongId)
		stream.Send(inp)
	}
	_, err = stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error while receiving response from LongGreet :%v\n", err)
	}
	//log.Println("Transactions result:", res)
}

func BroadcastNewEnd(client proto.BlockchainServiceClient) {
	//log.Println("BroadCastEnd was invoked")
	r, err := client.NewEnd(context.Background(), &proto.EndRequest{
		EndFlag: true,
	})
	if err != nil {
		log.Fatalf("Could not broadcast block: %v\n", err)
	}
	log.Println(r.ER)
}

func BroadCastNewTransaction(form, to string, amount int, fee int64) {
	id := test.SendRefName(form, to, amount, fee)
	for i := 0; i < len(addrsAlive); i++ {
		if addrsAlive[i] == constcoe.Address {
			continue
		}
		conn, err := grpc.Dial(addrsAlive[i], grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Failed to connect: %v\n", err)
		}
		defer conn.Close()
		client := proto.NewBlockchainServiceClient(conn)
		txp := blockchain.CreateTransactionPool()
		println("len:", len(txp.PubTx))
		var tx transaction.Transaction
		for _, t := range txp.PubTx {
			if bytes.Equal(t.ID, id) {
				tx = *t
				break
			}
		}
		NewTransaction(client, tx)
		NewInputs(client, tx)
		NewOutputs(client, tx)
		BroadcastNewEnd(client)
	}
	//blockchain.RemoveTransactionPoolFile()
}

func BroadCastAsset(client proto.BlockchainServiceClient) bool {
	//log.Println("BroadCastEnd was invoked")
	r, err := client.Asset(context.Background(), &proto.AssetRequest{
		Coins:         int64(test.BalanceRefName(constcoe.Refname, false)),
		WalletAddress: test.WalletAddress(constcoe.Refname),
	})
	if err != nil {
		log.Printf("Could not broadcast block: %v\n", err)
		return false
	}
	log.Println(r.AR)
	return true
}

func PrepareForNextRound() {
	for i := 0; i < len(addrsAlive); i++ {
		println("Prepare i:", i, " addr:", addrsAlive[i])
		conn, err := grpc.Dial(addrsAlive[i], grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("Failed to connect: %v\n", err)
			return
		}
		defer conn.Close()
		// 创建新客户端
		client := proto.NewBlockchainServiceClient(conn)
		// 先判断结点在线再广播
		BroadCastAsset(client)
	}
	//blockchain.RemoveTransactionPoolFile()
}

func MineForOneRound() {
	//挖矿太快会导致有时会结点挖矿并广播完成后另一个结点才开始ChooseCreator
	//now := time.Now().Unix()
	address := test.WalletAddress(constcoe.Refname)
	println("Mining...")
	block := test.Mine(address, utils.Address2PubHash([]byte(address)))
	// 这prehash传过来必出问题，print都报错，在test.Mine里print一点事都没有。而且hash无论在哪里print也都一点事没有，看了俩小时也不知道为什么只能改成用读写文件传递了
	prehash, readError := ioutil.ReadFile("the weirdest fucking bug.txt")
	if readError != nil {
		panic("sth wrong")
	}
	block.PrevHash = prehash
	//addrs := addresses.ReadAllAddress()
	addrsAlive = addresses.AlivePort()
	//if time.Now().Unix()-now <= 8 {
	//	time.Sleep(4 * time.Second)
	//}
	for i := 0; i < len(addrsAlive); i++ {
		conn, err := grpc.Dial(addrsAlive[i], grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("Failed to connect: %v\n", err)
		}
		defer conn.Close()
		// 创建新客户端
		client := proto.NewBlockchainServiceClient(conn)
		BB(client, block, addrsAlive[i])
	}
	test.ShowWallets(false)
	test.ShowWallets(true)
}

func StartClient() {
	addrsAlive = addresses.AlivePort()
	TestAlive()
	addrsAlive = addresses.AlivePort()
	BroadCastNewTransaction("G", "H", 17, 2)
	RedisInit()
	rdb.Set("winner", "nil", 0)
	PrepareForNextRound()
	for {
		winner, _ := rdb.Get("winner").Result()
		if winner == "nil" {
			continue
		}
		if winner == test.WalletAddress(constcoe.Refname) {
			MineForOneRound()
		}
		break
	}
}
