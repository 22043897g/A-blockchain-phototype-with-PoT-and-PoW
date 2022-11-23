package main

import (
	"BlockchainInGo/addresses"
	"BlockchainInGo/blockchain"
	"BlockchainInGo/constcoe"
	"BlockchainInGo/merkletree"
	"BlockchainInGo/myclient"
	"BlockchainInGo/proto"
	"BlockchainInGo/test"
	"BlockchainInGo/transaction"
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"github.com/go-redis/redis"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"log"
	"net"
	"strconv"
)

var rdb *redis.Client

func initRedis() (err error) {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379", // 指定
		Password: "",
		DB:       1, // redis一共16个库，指定其中一个库即可
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

type blockrequest struct {
	height  int64
	address string
}

type candidate struct {
	address    string
	balance    int64
	territoryL float64 //根据balance占全部候选人balance的比例和address顺序划分领地，当前块哈希前两位落在哪个结点的领地中该结点即为下一个块的生产者
	territoryR float64
}

var creator string
var block blockchain.Block
var newTX transaction.Transaction
var candidates []candidate

type Server struct {
	proto.BlockchainServiceServer
}

// Basic 组装接收到的区块时填入基础信息
func Basic(b *blockchain.Block, in *proto.BlockRequest) {
	b.Difficulty = in.Difficulty
	b.Timestamp = in.Timestamp
	b.Nonce = in.Nonce
	b.Target = in.Target
	b.Index = in.Index
	b.Hash = in.Hash
	b.PrevHash = in.PrevHash
	b.Creator = in.Creator
}

func (*Server) Alive(ctx context.Context, in *proto.AliveRequest) (*proto.AliveResponse, error) {
	log.Printf("Alive was invoked")
	return &proto.AliveResponse{Hi: constcoe.Address + "is online."}, nil
}

// GetBlock 节点请求获得所有区块，需要根据该结点最新区块高度决定发多少区块，若是新加入的结点还需记录其地址
func (*Server) GetBlock(ctx context.Context, in *proto.GetBlockRequest) (*proto.CreateByResponse, error) {
	// 查询该地址是否存在，不存在则记录
	if !addresses.CheckAddress(in.Address) {
		addresses.SaveNewAddress(in.Address)
	}
	// 将所有需要发送的区块放入blocks
	blocks := []blockchain.Block{}
	chain := blockchain.ContinueBlockChain()
	defer chain.Database.Close()
	iterator := chain.Iterator()
	//ogprevhash := chain.BackOgPrevHash()
	for {
		CHeight := blockchain.ReadLB().Index
		if in.Height >= CHeight {
			return &proto.CreateByResponse{WR: "No more blocks,the current height is " + strconv.FormatInt(CHeight, 10)}, nil
		}
		block := *iterator.Next()
		blocks = append(blocks, block)
		if block.Index == in.Height+1 {
			break
		}
	}
	println("================Blocks generated==================")
	for i := 0; i < len(blocks); i++ {
		println(blocks[i].Index)
	}
	//发送
	conn, err := grpc.Dial(in.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v\n", err)
	}
	defer conn.Close()
	client := proto.NewBlockchainServiceClient(conn)
	for i := len(blocks) - 1; i >= 0; i-- {
		if blocks[i].Index == 1 {
			myclient.BroadcastCreator(client, "A")
		}
		myclient.BB(client, block, in.Address)
	}
	return &proto.CreateByResponse{WR: "Blocks sent"}, nil
}

// CreateBy 接收创世区块生成者
func (*Server) CreateBy(ctx context.Context, in *proto.CreateByRequest) (*proto.CreateByResponse, error) {
	log.Printf("Createby was invoked with %s\n", in.Name)
	creator = in.Name
	return &proto.CreateByResponse{WR: "Received name: " + in.Name}, nil
}

// Block 接收区块
func (*Server) Block(ctx context.Context, in *proto.BlockRequest) (*proto.BlockResponse, error) {
	log.Printf("Transactions was invoked with %d\n", in.Index)
	Basic(&block, in)
	return &proto.BlockResponse{BR: "Received block" + strconv.FormatInt(in.Index, 10)}, nil
}

// Transactions 接收交易
func (*Server) Transactions(stream proto.BlockchainService_TransactionsServer) error {
	// 清空transaction
	txs := []*transaction.Transaction{}
	block.Transactions = txs
	log.Println("Transactions was invoked")
	count := int64(0)
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			res := "Transaction transmission finished," + strconv.FormatInt(count, 10) + " transactions have been received."
			return stream.SendAndClose(&proto.TxsResponse{
				TR: res,
			})
		}
		if err != nil {
			log.Fatalf("Error while reading client stream :%v\n", err)
		}
		// 注意这个输出在服务端
		count += 1
		//fmt.Printf("收到交易:%x,它属于%x\n", req.ID, req.BelongHash)
		tx := transaction.Transaction{
			ID:       req.ID,
			Inputs:   nil,
			TxOutput: nil,
			Fee:      req.Fee,
			Type:     req.Type,
			From:     req.From,
			TO:       req.To,
			Amount:   req.Amount,
		}
		block.Transactions = append(block.Transactions, &tx)
	}
}

// Inputs 接收交易输入
func (*Server) Inputs(stream proto.BlockchainService_InputsServer) error {
	log.Println("Inputs was invoked")
	count := int64(0)
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			res := "Inputs transmission finished," + strconv.FormatInt(count, 10) + " inputs have been received."
			return stream.SendAndClose(&proto.InputsResponse{
				IR: res,
			})
		}
		if err != nil {
			log.Fatalf("Error while reading client stream :%v\n", err)
		}
		// 注意这个输出在服务端
		count += 1
		//fmt.Printf("收到交易输入:%x,它属于%x\n", req.Index, req.BelongId)
		input := transaction.TxInput{
			TxID:   req.TxID,
			OutIdx: req.OutIdx,
			PubKey: req.PubKey,
			Sig:    req.Sig,
		}
		for i := 0; i < len(block.Transactions); i++ {
			tx := block.Transactions[i]
			if bytes.Equal(tx.ID, req.BelongId) {
				tx.Inputs = append(tx.Inputs, input)
			}
		}
	}
}

// Outputs 接收交易输出
func (*Server) Outputs(stream proto.BlockchainService_OutputsServer) error {
	log.Println("Outputs was invoked")
	count := int64(0)
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			res := "Inputs transmission finished," + strconv.FormatInt(count, 10) + " outputs have been received."
			return stream.SendAndClose(&proto.OutputsResponse{
				OR: res,
			})
		}
		if err != nil {
			log.Fatalf("Error while reading client stream :%v\n", err)
		}
		// 注意这个输出在服务端
		count += 1
		//fmt.Printf("收到交易输出:%x,它属于%x\n", req.Index, req.BelongId)
		output := transaction.TxOutput{
			Value:      int(req.Value),
			HashPubKey: req.HashPubKey,
		}
		for i := 0; i < len(block.Transactions); i++ {
			tx := block.Transactions[i]
			if bytes.Equal(tx.ID, req.BelongId) {
				tx.TxOutput = append(tx.TxOutput, output)
			}
		}
	}
}

// End 接收传送完毕信号，组装并验证新区块，如果验证成功则加入区块链
func (*Server) End(ctx context.Context, in *proto.EndRequest) (*proto.EndResponse, error) {
	block.Transactions = blockchain.Sort(block.Transactions)
	log.Printf("End was invoked")
	judge := ""
	if block.Index == 1 {
		test.CreateBlockChainRefName(creator, &block)
		judge = "passed."
	} else {
		root := merkletree.CreateMerkleTree(block.Transactions)
		block.MTree = root
		flag := VerifyBlock(block)
		if flag {
			judge = "passed."
			bc := blockchain.ContinueBlockChain()
			bc.AddBlock(&block)
			println("height", block.Index)
			blockchain.SaveLB(block)
			blockchain.SaveTime(block.Index)
		} else {
			judge = "failed."
		}
		blockchain.RemoveTransactionPoolFile()
		//txp := blockchain.CreateTransactionPool()
		//for _, tx := range block.Transactions {
		//	for _, txip := range txp.PubTx {
		//		if bytes.Equal(tx.ID, txip.ID) {
		//			txp.DeleteInvalidTransactions(txip)
		//		}
		//	}
		//}
		//fmt.Printf("Server_hash:%x", tools.VerifyHash(block))
		//fmt.Printf("Txs_hash:%x\n", tools.VerifyTxsHash(block))
	}
	test.ShowWallets(true)
	return &proto.EndResponse{
		ER: "Miner2:Block " + strconv.FormatInt(block.Index, 10) + " received,verification " + judge,
	}, nil
}

//以下接收的是新交易，收到后加入交易池

func (*Server) NewEnd(ctx context.Context, in *proto.EndRequest) (*proto.EndResponse, error) {
	log.Printf("NewEnd was invoked")
	tp := blockchain.CreateTransactionPool() //注意这个函数是创建空池并加载硬盘存的池
	tp.AddTransaction(&newTX)
	tp.SaveFile()
	println(tp.PubTx[0].From)
	println(tp.PubTx[0].TO)
	println(tp.PubTx[0].Amount)
	println(tp.PubTx[0].Fee)
	fmt.Println("Success!")

	return &proto.EndResponse{
		ER: "NewTransaction received",
	}, nil
}

func (*Server) NewTransaction(ctx context.Context, in *proto.TransactionsRequest) (*proto.TxsResponse, error) {
	log.Println("NewTransaction was invoked")
	tx := transaction.Transaction{
		ID:       in.ID,
		Inputs:   nil,
		TxOutput: nil,
		Fee:      in.Fee,
		Type:     in.Type,
		From:     in.From,
		TO:       in.To,
		Amount:   in.Amount,
	}
	newTX = tx

	return &proto.TxsResponse{TR: "Received new transaction"}, nil
}

func (*Server) NewOutPuts(stream proto.BlockchainService_NewOutPutsServer) error {
	log.Println("NewOutputs was invoked")
	count := int64(0)
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			res := "Inputs transmission finished," + strconv.FormatInt(count, 10) + " outputs have been received."
			return stream.SendAndClose(&proto.OutputsResponse{
				OR: res,
			})
		}
		if err != nil {
			log.Fatalf("Error while reading client stream :%v\n", err)
		}
		// 注意这个输出在服务端
		count += 1
		//fmt.Printf("收到交易输出:%x,它属于%x\n", req.Index, req.BelongId)
		output := transaction.TxOutput{
			Value:      int(req.Value),
			HashPubKey: req.HashPubKey,
		}
		newTX.TxOutput = append(newTX.TxOutput, output)
	}
}

func (*Server) NewInPuts(stream proto.BlockchainService_NewInPutsServer) error {
	log.Println("NewInputs was invoked")
	count := int64(0)
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			res := "Inputs transmission finished," + strconv.FormatInt(count, 10) + " inputs have been received."
			return stream.SendAndClose(&proto.InputsResponse{
				IR: res,
			})
		}
		if err != nil {
			log.Fatalf("Error while reading client stream :%v\n", err)
		}
		// 注意这个输出在服务端
		count += 1
		//fmt.Printf("收到交易输入:%x,它属于%x\n", req.Index, req.BelongId)
		input := transaction.TxInput{
			TxID:   req.TxID,
			OutIdx: req.OutIdx,
			PubKey: req.PubKey,
			Sig:    req.Sig,
		}
		newTX.Inputs = append(newTX.Inputs, input)
	}
}

func VerifyBlock(block blockchain.Block) bool {
	if !block.ValidPow() {
		return false
	} else {
		hash := block.CalculateHash()
		if !bytes.Equal(hash, block.Hash) {
			return false
		}
	}
	return true
}

// SortC Candidates按地址顺序从小到大排序
func SortC() {
	n := len(candidates)
	for i := n; i > 0; i-- {
		flag := false
		for j := 1; j < i; j++ {
			if candidates[j-1].address > candidates[j].address {
				candidates[j-1], candidates[j] = candidates[j], candidates[j-1]
				flag = true
			}
		}
		if flag == false {
			break
		}
	}
}

// ChooseCreator 选出下一轮的块创建者
func ChooseCreator() string {
	SortC()
	var L []float64
	var R []float64
	var P []float64
	Amount := float64(0)
	lastT := float64(0)
	for _, can := range candidates {
		Amount += float64(can.balance)
	}
	LH := hex.EncodeToString(blockchain.ReadLB().Hash)
	winner := ""
	if LH[62:] == "00" {
		return candidates[0].address
	}
	Cake, _ := strconv.ParseUint(LH[62:], 16, 32)
	for i, can := range candidates {
		var percentage float64
		percentage = float64(can.balance) / Amount
		P = append(P, percentage)
		can.territoryL = lastT
		L = append(L, can.territoryL)
		can.territoryR = lastT + percentage*255
		R = append(R, can.territoryR)
		println(Cake)
		fmt.Printf("i:%d addr:%s balance:%d percentage:%g L:%g R:%g\n", i, can.address, can.balance, percentage, can.territoryL, can.territoryR)
		if can.territoryL < float64(Cake) && float64(Cake) <= can.territoryR {
			winner = can.address
		}
		lastT = can.territoryR
	}
	if winner == test.WalletAddress(constcoe.Refname) {
		rdb = redis.NewClient(&redis.Options{
			Addr:     "127.0.0.1:6379", // 指定
			Password: "",
			DB:       10, // redis一共16个库，指定其中一个库即可
		})
		rdb.Set("Cake", Cake, 0)
		rdb.Set("Winner", winner, 0)
		for i := 0; i < len(candidates); i++ {
			key := strconv.FormatInt(int64(i), 10)
			b := strconv.FormatInt(candidates[i].balance, 10)
			p := strconv.FormatFloat(P[i], 'f', 5, 32)
			l := strconv.FormatFloat(L[i], 'f', 5, 32)
			r := strconv.FormatFloat(R[i], 'f', 5, 32)
			value := "addr:" + candidates[i].address + " balance:" + b + " percentage:" + p + " L:" + l + " R:" + r
			rdb.Set(key, value, 0)
		}

	}
	return winner
}

func (*Server) Asset(ctx context.Context, in *proto.AssetRequest) (*proto.AssetResponse, error) {
	log.Printf("Asset was invoked")
	can := candidate{address: in.WalletAddress, balance: in.Coins}
	if can.balance != 0 {
		candidates = append(candidates, can)
	}
	//myclient.TestAlive()
	println("CanNum:", len(candidates))
	println("PortsNum:", len(addresses.AlivePort()))
	if len(candidates) == len(addresses.AlivePort()) {
		winner := ChooseCreator()
		candidates = []candidate{}
		println("====================================winner:", winner)
		RedisInit()
		rdb.Set("winner", winner, 0)
	}
	return &proto.AssetResponse{AR: "Balance and address received"}, nil
}

func main() {
	lis, err := net.Listen("tcp", constcoe.Address)
	if err != nil {
		log.Fatalf("Failed to listen on: %v\n", err)
	}
	log.Printf("Listening on %s\n", constcoe.Address)
	//创建服务器
	server := grpc.NewServer()
	//注册(后面那个参数包含所有*Server接收器)
	proto.RegisterBlockchainServiceServer(server, &Server{})
	if err = server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve:%v\n", err)
	}
}
