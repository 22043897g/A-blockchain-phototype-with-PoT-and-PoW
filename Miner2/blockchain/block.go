package blockchain

import (
	"BlockchainInGo/constcoe"
	"BlockchainInGo/merkletree"
	"BlockchainInGo/transaction"
	"BlockchainInGo/utils"
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"github.com/go-redis/redis"
	"strconv"
	"time"
)

var rdb *redis.Client

type Block struct {
	Creator      string
	Index        int64 //高度
	Timestamp    int64
	Hash         []byte
	PrevHash     []byte
	Difficulty   int64
	Target       []byte //方便其他结点快速验证nonce是否正确
	Nonce        int64
	Transactions []*transaction.Transaction //记录交易信息
	MTree        *merkletree.MerkleTree
}

// SetHash 计算并设置区块hash
func (b *Block) SetHash() {
	// information 将区块的各属性连接起来，用以求哈希值，最后一个参数是连接时的分隔符，此处取空
	data := bytes.Join([][]byte{utils.Int64ToBytes(b.Index), utils.Int64ToBytes(b.Timestamp), b.PrevHash, b.Target, utils.Int64ToBytes(b.Nonce), b.BackTransactionSummary(), b.MTree.RootNode.Data}, []byte{})
	// sum256返回的是sha256的校验和，[32]byte形式，无法直接用b.Hash接收
	hash := sha256.Sum256(data)
	b.Hash = hash[:]
}

// CalculateHash 计算哈希，用于client发送时的测试
func (b *Block) CalculateHash() []byte {
	// information 将区块的各属性连接起来，用以求哈希值，最后一个参数是连接时的分隔符，此处取空
	data := bytes.Join([][]byte{utils.Int64ToBytes(b.Index), utils.Int64ToBytes(b.Timestamp), b.PrevHash, b.Target, utils.Int64ToBytes(b.Nonce), b.BackTransactionSummary(), b.MTree.RootNode.Data}, []byte{})
	// sum256返回的是sha256的校验和，[32]byte形式，无法直接用b.Hash接收
	hash := sha256.Sum256(data)
	return hash[:]
}

//func WriteLB(block Block) {
//	inputfile := "LB.txt"
//	outputfile := inputfile
//	bs, readError := ioutil.ReadFile(inputfile)
//	if readError != nil {
//		panic("sth wrong")
//	}
//	bs = block.Serialize()
//	// perm是读写权限
//	writeErr := ioutil.WriteFile(outputfile, bs, 0666)
//	if writeErr != nil {
//		panic("sth wrong")
//	}
//}

// SaveLB 存储最后一个区块的部分信息
func SaveLB(block Block) {
	RedisInit()
	timestamp := strconv.FormatInt(block.Timestamp, 10)
	difficulty := strconv.FormatInt(block.Difficulty, 10)
	index := strconv.FormatInt(block.Index, 10)
	hash := hex.EncodeToString(block.Hash)
	prevhash := hex.EncodeToString(block.PrevHash)
	value := "index:" + index + " difficulty:" + difficulty + " hash:" + hash + " prevhash:" + prevhash + " timestamp:" + timestamp
	key := "B" + index
	rdb.Set(key, value, 0)
	rdb.Set("difficulty", difficulty, 0)
	rdb.Set("index", index, 0)
	rdb.Set("hash", hash, 0)
	rdb.Set("prevhash", prevhash, 0)
}

// ReadLB 读取difficulty和index
func ReadLB() Block {
	RedisInit()
	block := Block{}
	res, _ := rdb.Get("difficulty").Result()
	difficulty, _ := strconv.ParseInt(res, 10, 64)
	res2, _ := rdb.Get("index").Result()
	index, _ := strconv.ParseInt(res2, 10, 64)
	res3, _ := rdb.Get("hash").Result()
	hash, _ := hex.DecodeString(res3)
	block.Index = index
	block.Difficulty = difficulty
	block.Hash = hash
	return block
}

// redis初始化
func initRedis() (err error) {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379", // 指定
		Password: "",
		DB:       0, // redis一共16个库，指定其中一个库即可
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

// SaveTime and ReadTime   存写时间，用于计算interval以决定difficulty
func SaveTime(index int64) {
	RedisInit()
	now := time.Now().Unix()
	si := strconv.FormatInt(index, 10)
	sn := strconv.FormatInt(now, 10)
	rdb.Set(si, sn, 0)
}

func ReadTime() []int64 {
	RedisInit()
	var times []int64
	key := ReadLB().Index - 9
	for i := 0; i < 10; i++ {
		res, _ := rdb.Get(strconv.FormatInt(key, 10)).Result()
		t, _ := strconv.ParseInt(res, 10, 64)
		times = append(times, t)
		key++
	}
	for i := 0; i < 10; i++ {
		println("i:", i, "time:", times[i])
	}
	return times
}

func Sort(txs []*transaction.Transaction) []*transaction.Transaction {
	n := len(txs)
	for i := n; i > 0; i-- {
		flag := false
		for j := 1; j < i; j++ {
			if hex.EncodeToString(txs[j-1].ID) > hex.EncodeToString(txs[j].ID) {
				txs[j-1], txs[j] = txs[j], txs[j-1]
				flag = true
			}
		}
		if flag == false {
			break
		}
	}
	return txs
}

// CreateBlock 创建区块
func CreateBlock(addr string, prevhash []byte, txs []*transaction.Transaction) *Block {
	var block Block
	lastBlock := ReadLB()
	txs = Sort(txs)
	if bytes.Equal(prevhash, []byte("And there was light.")) {
		block = Block{addr, 1, time.Now().Unix(), []byte{}, prevhash, constcoe.InitDifficulty, []byte{}, 0, txs, merkletree.CreateMerkleTree(txs)}
		SaveLB(block)
	} else {
		height := lastBlock.Index
		block = Block{addr, height + 1, time.Now().Unix(), []byte{}, prevhash, 0, []byte{}, 0, txs, merkletree.CreateMerkleTree(txs)}
	}
	block.Target = block.GetTarget()
	block.Nonce = block.FindNonce()
	block.SetHash()
	return &block
}

// GenesisBlock 创建第一笔交易，并将其放入创世区块
func GenesisBlock(addr string) *Block {
	tx := transaction.BaseTx(addr, constcoe.InitCoin, true)
	genesis := CreateBlock(hex.EncodeToString(utils.Address2PubHash([]byte(addr))), []byte("And there was light."), []*transaction.Transaction{tx})
	genesis.SetHash()
	return genesis
}

// BackTransactionSummary 将所有交易信息整合为一个byte slice,以帮助SetHash函数和GetBase4Nonce函数进行序列化
func (b *Block) BackTransactionSummary() []byte {
	txIDs := make([][]byte, 0)
	for _, tx := range b.Transactions {
		txIDs = append(txIDs, tx.ID)
	}
	summary := bytes.Join(txIDs, []byte{})
	return summary
}

// ValidPow 计算hash并与target比较以验证nonce是否合法(注意需要计算hash不能直接使用区块中的hash）
func (b *Block) ValidPow() bool {
	targer := utils.BytesToInt64(b.Target)
	data := b.GetBase4Nonce(b.Nonce)
	hash := sha256.Sum256(data)
	hash_b := hash[:]
	hash_64 := utils.BytesToInt64(hash_b)
	if hash_64 <= targer {
		return true
	} else {
		return false
	}
}

// Badger只能序列化存储,所以要有序列化和反序列化函数

// Serialize 将数据序列化
func (b *Block) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)
	err := encoder.Encode(b)
	utils.Handle(err)
	return res.Bytes()
}

// DeSerialize 反序列化
func DeSerialize(data []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&block)
	utils.Handle(err)
	return &block
}
