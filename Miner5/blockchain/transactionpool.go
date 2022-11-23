package blockchain

import (
	"BlockchainInGo/constcoe"
	"BlockchainInGo/transaction"
	"BlockchainInGo/utils"
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"
)

// TransactionPool 交易池，用于存储交易信息
type TransactionPool struct {
	PubTx []*transaction.Transaction
}

// DeleteInvalidTransactions 删除无效（双花）交易，其实是创建新交易池并删除旧的
func (tp *TransactionPool) DeleteInvalidTransactions(tx *transaction.Transaction) {
	newPubTx := TransactionPool{}
	for _, t := range tp.PubTx {
		if !bytes.Equal(t.ID, tx.ID) {
			newPubTx.PubTx = append(newPubTx.PubTx, t)
		}
	}
	err := RemoveTransactionPoolFile()
	if err != nil {
		utils.Handle(err)
	}
	newPubTx.SaveFile()
}

// AddTransaction 添加交易
func (tp *TransactionPool) AddTransaction(tx *transaction.Transaction) {
	tp.PubTx = append(tp.PubTx, tx)
}

// SaveFile 将交易保存到磁盘
func (tp *TransactionPool) SaveFile() {
	var content bytes.Buffer
	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(tp)
	utils.Handle(err)
	err = ioutil.WriteFile(constcoe.TransactionPoolFile, content.Bytes(), 0644)
	utils.Handle(err)
}

// LoadFile 将磁盘中的数据读到交易池
func (tp *TransactionPool) LoadFile() error {
	if !utils.FileExists(constcoe.TransactionPoolFile) {
		return nil
	}
	var transactionPool TransactionPool
	fileContent, err := ioutil.ReadFile(constcoe.TransactionPoolFile)
	if err != nil {
		return err
	}

	decoder := gob.NewDecoder(bytes.NewBuffer(fileContent))
	err = decoder.Decode(&transactionPool)
	if err != nil {
		return err
	}

	tp.PubTx = transactionPool.PubTx
	return nil
}

// CreateTransactionPool 加载或创建空交易池
func CreateTransactionPool() *TransactionPool {
	TransactionPool := TransactionPool{}
	err := TransactionPool.LoadFile()
	utils.Handle(err)
	return &TransactionPool
}

// RemoveTransactionPoolFile 挖矿结束后删除交易池
func RemoveTransactionPoolFile() error {
	err := os.Remove(constcoe.TransactionPoolFile)
	if err == nil {
		fmt.Println("TxPool removed")
	}
	return err
}
