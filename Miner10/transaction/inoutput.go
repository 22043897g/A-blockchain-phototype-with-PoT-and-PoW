package transaction

import (
	"BlockchainInGo/utils"
	"bytes"
)

// 以下两者均为结构体Transaction的一部分

// TxOutput 用于表明钱去哪
type TxOutput struct {
	Value      int    //转出资产值
	HashPubKey []byte //转出地址（钱转给谁），注意此处用的地址哈希值以保护隐私
}

// TxInput 用于表明钱从哪来
type TxInput struct {
	TxID   []byte // 前置交易ID
	OutIdx int64  // 前置交易中的第几个Output(每个交易有多个交易输出且前置交易的交易输出是本交易的交易输入)
	PubKey []byte // 转入者地址(钱从哪里转来)
	Sig    []byte //签名
}

// FromAddressRight 以下两函数用于比较地址是否正确（一个交易中有多个交易输入与输出，需要遍历并判断其地址是否与要输入/输出地址一致）
func (in *TxInput) FromAddressRight(address []byte) bool {
	return bytes.Equal(in.PubKey, address)
}

func (in *TxOutput) ToAddressRight(address []byte) bool {
	return bytes.Equal(in.HashPubKey, utils.PublicKeyHash(address))
}
