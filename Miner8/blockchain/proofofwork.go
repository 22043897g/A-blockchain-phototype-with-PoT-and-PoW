package blockchain

import (
	"BlockchainInGo/utils"
	"bytes"
	"crypto/sha256"
	"math"
)

// GetTarget 通过常量Difficulty设置target
func (b *Block) GetTarget() []byte {
	var difficulty int64
	if bytes.Equal(b.PrevHash, []byte("And there was light.")) {
		difficulty = b.Difficulty
	} else {
		lastblock := ReadLB()
		difficulty = lastblock.Difficulty
	}
	var target int64
	target = math.MaxInt64
	// 右移difficulty位
	target = target >> difficulty
	return utils.Int64ToBytes(target)
}

// GetBase4Nonce 整合所有数据，用以计算哈希并确定是否满足target
func (b *Block) GetBase4Nonce(nonce int64) []byte {
	data := bytes.Join([][]byte{
		utils.Int64ToBytes(b.Index),
		utils.Int64ToBytes(b.Timestamp),
		b.PrevHash,
		b.Target,
		utils.Int64ToBytes(int64(nonce)),
		b.BackTransactionSummary(),
		b.MTree.RootNode.Data,
	},
		[]byte{},
	)
	return data
}

// FindNonce 遍历以找到Nonce
func (b *Block) FindNonce() int64 {
	var hash_b []byte
	var target int64
	var nonce int64
	nonce = 0
	target = utils.BytesToInt64(b.Target)

	for nonce < math.MaxInt64 {
		data := b.GetBase4Nonce(nonce)
		hash := sha256.Sum256(data)
		hash_b = hash[:]
		hash_64 := utils.BytesToInt64(hash_b)
		// ==-1 即intHash<intTarget，即符合条件
		if hash_64 < target {
			break
		} else {
			nonce++
		}
	}
	return nonce
}
