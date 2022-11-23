package test

import (
	"BlockchainInGo/blockchain"
	"BlockchainInGo/utils"
	"BlockchainInGo/wallet"
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/go-redis/redis"
	"io/ioutil"
	"strconv"
)

var rdb *redis.Client

// redis初始化
func initRedis() (err error) {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379", // 指定
		Password: "",
		DB:       11, // redis一共16个库，指定其中一个库即可
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

// SendRefName 根据钱包别名打钱
func SendRefName(from, to string, amount int, fee int64) []byte {
	refList := wallet.LoadRefList()
	fromAddress, err := refList.FindRef(from)
	utils.Handle(err)
	toAddress, err := refList.FindRef(to)
	utils.Handle(err)
	return Send(fromAddress, toAddress, amount, fee)
}

func Send(from, to string, amount int, fee int64) []byte {
	chain := blockchain.ContinueBlockChain()
	defer chain.Database.Close()
	fromWallet := wallet.LoadWallet(from)
	tx, ok := chain.CreateTransaction(fromWallet.PublicKey, utils.Address2PubHash([]byte(to)), amount, int(fee), fromWallet.PrivateKey)
	tx.From = from
	tx.TO = to
	tx.Fee = fee
	tx.Amount = int64(amount)
	if !ok {
		fmt.Println("Failed to create transaction")
		return nil
	}
	tp := blockchain.CreateTransactionPool() //注意这个函数是创建空池并加载硬盘存的池
	tp.AddTransaction(tx)
	tp.SaveFile()
	fmt.Println("Success!")
	return tx.ID
}

func CreateGenBlock(address string) *blockchain.Block {
	Gen := blockchain.GenesisBlock(address)
	fmt.Println("Finished creating GenBlock")
	return Gen
}

// CreateGenBlockRefName 根据钱包别名创建创世区块
func CreateGenBlockRefName(refname string) *blockchain.Block {
	refList := wallet.LoadRefList()
	address, err := refList.FindRef(refname)
	utils.Handle(err)
	block := CreateGenBlock(address)
	return block
}

// CreateBlockChainRefName 根据钱包别名创建区块链
func CreateBlockChainRefName(refname string, block *blockchain.Block) {
	refList := wallet.LoadRefList()
	address, err := refList.FindRef(refname)
	utils.Handle(err)
	CreateBlockChain(address, block)
}

func CreateBlockChain(address string, block *blockchain.Block) {
	newChain := blockchain.InitBlockChain(utils.Address2PubHash([]byte(address)), block)
	newChain.Database.Close()
	fmt.Println("Finished creating blockchain,and the owner is :", address)
}

func Mine(addr string, address []byte) blockchain.Block {
	chain := blockchain.ContinueBlockChain()
	defer chain.Database.Close()
	block := chain.RunMine(addr, address)
	writeErr := ioutil.WriteFile("the weirdest bug.txt", block.PrevHash, 0666)
	if writeErr != nil {
		panic("sth wrong")
	}
	//println("---------------------------------------3---------------------------------------")
	//fmt.Printf("hash:%x\n", block.Hash)
	//fmt.Println(block.PrevHash)
	//println("---------------------------------------4---------------------------------------")
	fmt.Println("Finish Mining")
	return block
}

func BlockChainInfo() {
	chain := blockchain.ContinueBlockChain()
	defer chain.Database.Close()
	iterator := chain.Iterator()
	ogprevhash := chain.BackOgPrevHash()
	for {
		block := iterator.Next()
		fmt.Println("--------------------------------------------------------------------------------------------------------------")
		fmt.Printf("Creator:%s\n", block.Creator)
		fmt.Printf("Difficulty:%d\n", block.Difficulty)
		fmt.Printf("Height:%d\n", block.Index)
		fmt.Printf("Timestamp:%d\n", block.Timestamp)
		fmt.Printf("Previous hash:%x\n", block.PrevHash)
		fmt.Printf("Number of transactions:%d\n", len(block.Transactions))
		fmt.Printf("hash:%x\n", block.Hash)
		fmt.Println("--------------------------------------------------------------------------------------------------------------")
		fmt.Println()
		if bytes.Equal(block.PrevHash, ogprevhash) {
			break
		}
	}
}

func Balance(address string, Type bool, name string) int {
	RedisInit()
	chain := blockchain.ContinueBlockChain() //读出区块链
	defer chain.Database.Close()
	wlt := wallet.LoadWallet(address)
	balance, _ := chain.FindUTXOs(wlt.PublicKey, Type)
	value := "Address:" + address + " balance:" + strconv.FormatInt(int64(balance), 10)
	if Type == false {
		rdb.Set(name+"0", value, 0)
	}
	if Type == true {
		rdb.Set(name+"1", value, 0)
	}
	fmt.Printf("Address:%s,balance:%d\n", address, balance)
	return balance
}

func BalanceRefName(refname string, Type bool) int {
	refList := wallet.LoadRefList()
	address, err := refList.FindRef(refname)
	utils.Handle(err)
	println("===================Name:", refname, "=========================")
	return Balance(address, Type, refname)
}

func WalletAddress(refname string) string {
	refList := wallet.LoadRefList()
	address, err := refList.FindRef(refname)
	utils.Handle(err)
	return address
}

func WalletsList() {
	refList := wallet.LoadRefList()
	for address, _ := range *refList {
		wlt := wallet.LoadWallet(address)
		fmt.Println("------------------------------------------------------")
		fmt.Printf("Wallet address:%x\n\n", address)
		fmt.Printf("Public Key:%x\n", wlt.PrivateKey)
		fmt.Printf("Reference Name:%s\n", (*refList)[address])
		fmt.Println("------------------------------------------------------")
		fmt.Println()
	}
}

func ShowWallets(Type bool) {
	reflist := *wallet.LoadRefList()
	for _, v := range reflist {
		BalanceRefName(v, Type)
	}
}

func GetPrivateKet(name string) string {
	wlt := wallet.LoadWallet(WalletAddress("A"))
	return hex.EncodeToString(wlt.PrivateKey.D.Bytes())
}
