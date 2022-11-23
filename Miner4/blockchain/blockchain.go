//blockchain.go
package blockchain

import (
	"BlockchainInGo/constcoe"
	"BlockchainInGo/transaction"
	"BlockchainInGo/utils"
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"github.com/dgraph-io/badger"
	"runtime"
)

// BlockChain 只记录最后一个块的哈希值和存在数据库中的指针
type BlockChain struct {
	LastHash []byte
	Database *badger.DB
}

// BlockChainIterator 区块链迭代器，用于遍历区块链
type BlockChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

// AddBlock 在区块链中添加新的区块
func (bc *BlockChain) AddBlock(newBlock *Block) {
	var lastHash []byte

	err := bc.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		utils.Handle(err)
		err = item.Value(func(val []byte) error {
			lastHash = val
			return nil
		})
		utils.Handle(err)

		return err
	})
	utils.Handle(err)
	if !bytes.Equal(newBlock.PrevHash, lastHash) {
		fmt.Println("This block is out of age")
		runtime.Goexit()
	}

	err = bc.Database.Update(func(transaction *badger.Txn) error {
		err := transaction.Set(newBlock.Hash, newBlock.Serialize()) //将新区块写入数据库
		utils.Handle(err)
		err = transaction.Set([]byte("lh"), newBlock.Hash) // 更新最后一个区块的哈希值
		bc.LastHash = newBlock.Hash
		return err
	})
	defer bc.Database.Close()
	utils.Handle(err)
	fmt.Println("New block added.")
}

// func CreateBlockChain() *BlockChain {
// 	blockchain := BlockChain{}
// 	blockchain.Blocks = append(blockchain.Blocks, GenesisBlock())
// 	return &blockchain
// }

// InitBlockChain 初始化区块链
//使用bedger存储区块链，bedger只能存键值对
func InitBlockChain(address []byte, block *Block) *BlockChain {
	var lastHash []byte

	if utils.FileExists(constcoe.BCFile) {
		fmt.Println("blockchain already exists")
		// 终止当前协程,相对更安全，用普通退出也可以
		runtime.Goexit()
	}
	// 使用badger的默认配置
	opts := badger.DefaultOptions(constcoe.BCPath)
	opts.Logger = nil
	//数据库中的操作不输出到标准输出中
	db, err := badger.Open(opts)
	utils.Handle(err)
	//更新数据库，即向数据库中存入数据，注意这个Update函数的参数是个函数，为了方便创建事务
	err = db.Update(func(txn *badger.Txn) error {
		genesis := block
		fmt.Println("Genesis Created ")
		err = txn.Set(genesis.Hash, genesis.Serialize()) //存创世区块哈希值
		utils.Handle(err)
		err = txn.Set([]byte("lh"), genesis.Hash) //存当前区块链最后一个区块的哈希值
		utils.Handle(err)
		err = txn.Set([]byte("ogprevhash"), genesis.PrevHash) //存储前哈希值
		utils.Handle(err)
		lastHash = genesis.Hash
		return err
	})
	utils.Handle(err)
	blockchain := BlockChain{lastHash, db}
	//记录创建块的时间
	//utils.WriteBlockTime(strconv.FormatInt(time.Now().Unix(), 10))
	//utils.WriteBlockTime("\n")
	return &blockchain
}

// ContinueBlockChain 读取最后一个区块的哈希值
func ContinueBlockChain() *BlockChain {
	if utils.FileExists(constcoe.BCFile) == false {
		fmt.Println("No blockchain found, please create one first")
		runtime.Goexit()
	}

	var lastHash []byte

	opts := badger.DefaultOptions(constcoe.BCPath)
	opts.Logger = nil
	db, err := badger.Open(opts)
	utils.Handle(err)

	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		utils.Handle(err)
		err = item.Value(func(val []byte) error {
			lastHash = val
			return nil
		})
		utils.Handle(err)
		return err
	})
	utils.Handle(err)

	chain := BlockChain{lastHash, db}
	return &chain
}

// BackOgPrevHash 找到创世区块的前哈希值（用于与迭代器的currenthash比较以判断是否迭代到了创世区块）
func (chain *BlockChain) BackOgPrevHash() []byte {
	var ogprevhash []byte
	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("ogprevhash"))
		utils.Handle(err)

		err = item.Value(func(val []byte) error {
			ogprevhash = val
			return nil
		})

		utils.Handle(err)
		return err
	})
	utils.Handle(err)

	return ogprevhash
}

// Iterator 初始化迭代器
func (chain *BlockChain) Iterator() *BlockChainIterator {
	iterator := BlockChainIterator{chain.LastHash, chain.Database}
	return &iterator
}

// Next 向前遍历，每次返回一个BLOCK
func (iterator *BlockChainIterator) Next() *Block {
	var block *Block

	err := iterator.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iterator.CurrentHash)
		utils.Handle(err)

		err = item.Value(func(val []byte) error {
			block = DeSerialize(val)
			return nil
		})
		utils.Handle(err)
		return err
	})
	utils.Handle(err)

	iterator.CurrentHash = block.PrevHash

	return block
}

// FindUnspentTransactions 在整条区块链上寻找某个地址对应的包含该地址UTXO的交易
// 大前提是对于一笔交易，其有多个交易输入和多个交易输出，且可能多个交易输入的FromAddress是一个人，同时，交易输出中可能也有这个人（给自己找零）
func (bc *BlockChain) FindUnspentTransactions(address []byte, Type bool) []transaction.Transaction {
	var unSpentTxs []transaction.Transaction
	SpentTxs := make(map[string][]int64) //记录的是前置交易序号和第几个交易输出，key:交易ID,value:output在该交易中的序号（注意是一组，因为这些钱即使来自同一个人也可能是多笔钱）
	iter := bc.Iterator()
	//  注意[]byte不能作为key，所以要转化为string
	// 从尾到头遍历所有区块 （注意，区块链中有很多区块，每个区块中有多个交易，每个交易有多个交易输出，所以是三重循环）
all:
	for {
		// 由于[]byte不能作为key，需将交易ID转化为string
		block := iter.Next()
		//遍历当前区块中记录的所有交易
		for _, tx := range block.Transactions {
			if tx.Type == Type {
				txID := hex.EncodeToString(tx.ID)

			IterOutputs:
				// 遍历当前交易的所有输出
				for outIdx, out := range tx.TxOutput {
					// 此if的含义为“已花费”中有此交易的某个交易输出
					// 注意一个交易有多个交易输出，且SpentTxs的value是一组序号，代表某个交易中的多个交易输出是否已经被收款者花费
					// 故当某个交易输出的序号已被记录时（spentOut == outIdx）说明其已花费，故可以直接跳过
					// 但存在当前交易被记录，而该交易中的某个交易输出未被记录的情况，说明这部分交易输出并未被收款者花费
					// 此时无法找到spentOut == outIdx，不会continue，而是转入下一个if
					if SpentTxs[txID] != nil {
						for _, spentOut := range SpentTxs[txID] {
							if spentOut == int64(outIdx) {
								// 跳过当前交易输出
								continue IterOutputs
							}
						}
					}
					//当“已花费”中没有此交易或“已花费”中有此交易，但有可能会找零给自己
					if out.ToAddressRight(address) {
						unSpentTxs = append(unSpentTxs, *tx)
					}
				}
				// 检查所有非创建奖励（创建奖励没有交易输入，故无法通过检查交易输入判断资金是否花出）
				if !tx.IsBase() {
					for _, in := range tx.Inputs {
						//若某条交易中的输入地址为当前所检查的地址，则说明该地址进行了资金输出，故将其添加到已花费中
						if in.FromAddressRight(address) {
							inTxID := hex.EncodeToString(in.TxID)
							// 注意in是结构体Inputs的实例，故in.OutIdx表明前置交易的输出，即当前交易的输入。故此代码含义为别人给“我”的钱已经花出去了
							SpentTxs[inTxID] = append(SpentTxs[inTxID], in.OutIdx)
						}
					}
				}
			}
		}
		// 如果已经遍历完创世区块则break
		if bytes.Equal(block.PrevHash, bc.BackOgPrevHash()) {
			break all
		}

	}
	return unSpentTxs
}

// FindSpendableOutputs  找到某地址的部分UTXO即其总金额（这些UTXO的总金额需大于想花费的金额）（UTXO以交易ID和交易输出序号表示）
func (bc *BlockChain) FindSpendableOutputs(address []byte, amount int, Type bool) (int, map[string]int) {
	unspentOuts := make(map[string]int)
	// 找到当前地址对应的所有包含UTXO的交易
	unspentTxs := bc.FindUnspentTransactions(address, Type)
	accumulated := 0 // 已遍历UTXO的总金额
	// 因为交易中对于某个地址只可能有一个交易输出，所以当找到所需地址的交易输出时需直接continue外层循环以遍历下一个交易
Work:
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)
		for outIdx, out := range tx.TxOutput {
			if out.ToAddressRight(address) && accumulated < amount {
				accumulated += out.Value
				//记录UTXO（实际上记录的是交易的ID和交易输出序号）
				unspentOuts[txID] = outIdx
				if accumulated >= amount {
					break Work
				}
				continue Work // one transaction can only have one output referred to adderss
			}
		}
	}
	return accumulated, unspentOuts
}

// FindUTXOs 找到某地址的所有UTXO即其总金额（UTXO以交易ID和交易输出序号表示）
func (bc *BlockChain) FindUTXOs(address []byte, Type bool) (int, map[string]int) {
	unspentOuts := make(map[string]int)
	// 找到当前地址对应的所有包含UTXO的交易
	unspentTxs := bc.FindUnspentTransactions(address, Type)
	accumulated := 0
	// 因为交易中对于某个地址只可能有一个交易输出，所以当找到所需地址的交易输出时需直接continue外层循环以遍历下一个交易
Work:
	for _, tx := range unspentTxs {
		if tx.Type == Type {
			txID := hex.EncodeToString(tx.ID)
			for outIdx, out := range tx.TxOutput {
				if out.ToAddressRight(address) {
					accumulated += out.Value
					//记录UTXO（实际上记录的是交易的ID和交易输出序号）
					unspentOuts[txID] = outIdx
					continue Work
				}
			}
		}
	}
	return accumulated, unspentOuts
}

// CreateTransaction 进行消费，若UTXO>消费金额则创建新交易并返回true，否则不创建并返回false
func (bc *BlockChain) CreateTransaction(from_PubKey, to_HashPubKey []byte, amount int, fee int, privkey ecdsa.PrivateKey) (*transaction.Transaction, bool) {
	var inputs []transaction.TxInput                                        //交易输入
	var outputs []transaction.TxOutput                                      //交易输出
	acc, validOutputs := bc.FindSpendableOutputs(from_PubKey, amount, true) // 找到满足消费金额的所有UTXO及其总金额
	println("===========CreateTransaction==========")
	for i, out := range outputs {
		fmt.Printf("i:%d out:%d\n", i, out.Value)
	}
	// 金额不足时不创建任何新交易并返回false
	if acc < amount {
		fmt.Println("Not enough coins!")
		return &transaction.Transaction{}, false
	}
	for txid, outidx := range validOutputs {
		txID, err := hex.DecodeString(txid)
		utils.Handle(err)
		input := transaction.TxInput{txID, int64(outidx), from_PubKey, nil}
		inputs = append(inputs, input)
	}
	println("=================================CreateTransaction1:acc:", acc, " amount:", amount, " fee", fee)
	outputs = append(outputs, transaction.TxOutput{amount - fee, to_HashPubKey})
	//交易费
	outputs = append(outputs, transaction.TxOutput{fee, utils.Int64ToBytes(0)})
	if acc > amount {
		//找零
		outputs = append(outputs, transaction.TxOutput{acc - amount, utils.PublicKeyHash(from_PubKey)})
	}
	tx := transaction.Transaction{nil, inputs, outputs, 0, true, "", "", 0}
	tx.SetID()
	tx.Sign(privkey)
	return &tx, true
}

func BlockchainExist() bool {
	if utils.FileExists(constcoe.BCFile) == false {
		fmt.Println("No blockchain found, please create one first")
		return false
	} else {
		return true
	}
}

func Showpool() {
	txp := *CreateTransactionPool()
	for i, tx := range txp.PubTx {
		fmt.Printf("i:%d ID:%x From:%s To:%s Amount:%d Fee:%d\n ", i, tx.ID, tx.From, tx.TO, tx.Amount, tx.Fee)
		for i, ot := range tx.TxOutput {
			fmt.Printf("i:%d Value:%d\n", i, ot.Value)
		}
	}
}
