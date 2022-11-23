package transaction

import (
	"BlockchainInGo/utils"
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/gob"
	"math/rand"
	"strconv"
	"time"
)

type Transaction struct {
	ID       []byte     //本次交易ID(哈希值）
	Inputs   []TxInput  //一组输入 (钱从哪来)
	TxOutput []TxOutput //一组输出（钱去哪，包含输出给自身以实现找零）
	Fee      int64      //交易费
	Type     bool       //True为TXcoin（用于交易）,False为Feecoin（用于选出区块生产者）
	From     string
	TO       string
	Amount   int64
}

// 求交易的ID值(序列化后求哈希)
func (tx *Transaction) TxHash() []byte {
	var encoded bytes.Buffer
	var hash [32]byte

	// 编码为二进制格式
	encoder := gob.NewEncoder(&encoded)
	err := encoder.Encode(tx)
	utils.Handle(err)
	hash = sha256.Sum256(encoded.Bytes())
	return hash[:]
}

func (tx *Transaction) SetID() {
	tx.ID = tx.TxHash()
}

// BaseTx 给区块创建者奖励，toaddress即为区块创建者地址
func BaseTx(to string, value int64, Type bool) *Transaction {
	txIn := TxInput{[]byte{}, -1, []byte{}, nil} // 创建区块时的奖励币显然不需要"从何而来"
	TxOut := TxOutput{int(value), utils.Address2PubHash([]byte(to))}
	idnonce := strconv.FormatInt(rand.Int63n(10000000000000), 10) + strconv.FormatInt(time.Now().Unix(), 10)
	id := sha256.Sum256([]byte(idnonce))
	tx := Transaction{id[:], []TxInput{txIn}, []TxOutput{TxOut}, 0, Type, "BaseTX", to, int64(value)}
	return &tx
}

// IsBase 判断某交易是否为区块创建奖励
func (tx Transaction) IsBase() bool {
	return len(tx.Inputs) == 1 && tx.Inputs[0].OutIdx == -1
}

// PlainCopy 制作一个不带签名的交易，用以交易签名
func (tx *Transaction) PlainCopy() Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	for _, txin := range tx.Inputs {
		inputs = append(inputs, TxInput{txin.TxID, txin.OutIdx, nil, nil})
	}

	for _, txout := range tx.TxOutput {
		outputs = append(outputs, TxOutput{txout.Value, txout.HashPubKey})
	}

	txCopy := Transaction{tx.ID, inputs, outputs, tx.Fee, tx.Type, tx.From, tx.TO, tx.Amount}
	return txCopy
}

// PlainHash 通过外层循环给每个交易输入添加转入者地址
func (tx *Transaction) PlainHash(inidx int, prevPubKey []byte) []byte {
	txCopy := tx.PlainCopy()
	txCopy.Inputs[inidx].PubKey = prevPubKey
	return txCopy.TxHash()
}

// Sign 交易签名
func (tx *Transaction) Sign(privkey ecdsa.PrivateKey) {
	if tx.IsBase() {
		return
	}
	// 对每一个交易输入签名
	for idx, input := range tx.Inputs {
		plainhash := tx.PlainHash(idx, input.PubKey) //plainhash中除了签名以外其他信息都是全的
		signature := utils.Sign(plainhash, privkey)  //根据这些信息生成签名
		tx.Inputs[idx].Sig = signature               //签名
	}
}

// Verify 验证交易
func (tx *Transaction) Verify() bool {
	for idx, input := range tx.Inputs {
		plainhash := tx.PlainHash(idx, input.PubKey)
		if !utils.VerifySig(plainhash, input.PubKey, input.Sig) {
			return false
		}
	}
	return true
}
