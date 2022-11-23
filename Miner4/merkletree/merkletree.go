package merkletree

import (
	"BlockchainInGo/transaction"
	"BlockchainInGo/utils"
	"bytes"
	"crypto/sha256"
	"errors"
)

type MerkleTree struct {
	RootNode *MerkleNode
}

type MerkleNode struct {
	LeftNode  *MerkleNode
	RightNode *MerkleNode
	Data      []byte
}

// CreateMerkleNode 创建新树节点
func CreateMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	tempNode := MerkleNode{}
	if left == nil && right == nil {
		tempNode.Data = data
	} else {
		catenateHash := append(left.Data, right.Data...)
		hash := sha256.Sum256(catenateHash)
		tempNode.Data = hash[:]
	}

	tempNode.LeftNode = left
	tempNode.RightNode = right
	return &tempNode
}

// CreateMerkleTree 创建merkle树
func CreateMerkleTree(txs []*transaction.Transaction) *MerkleTree {
	txslen := len(txs)
	if txslen == 0 {
		node := CreateMerkleNode(nil, nil, utils.Int64ToBytes(0))
		MerkleTree := MerkleTree{node}
		return &MerkleTree
	}
	// 交易数为奇数时，复制最后一个交易，方便构建树
	if txslen%2 != 0 {
		txs = append(txs, txs[txslen-1])
	}

	//叶子结点存储
	var nodePool []*MerkleNode
	for _, tx := range txs {
		nodePool = append(nodePool, CreateMerkleNode(nil, nil, tx.ID))
	}
	for len(nodePool) > 1 {
		var tempNodePool []*MerkleNode
		poolLen := len(nodePool)
		if poolLen%2 != 0 { //某层出现奇数结点时，直接将该结点加入上一层（即不考虑兄弟结点哈希值）
			tempNodePool = append(tempNodePool, nodePool[poolLen-1])
		}
		for i := 0; i < poolLen/2; i++ {
			tempNodePool = append(tempNodePool, CreateMerkleNode(nodePool[2*i], nodePool[2*i+1], nil))
		}
		nodePool = tempNodePool //这步很重要，画个图就能理解
	}
	MerkleTree := MerkleTree{nodePool[0]}
	return &MerkleTree
}

//Find 深度优先遍历，一直递归到最左下，如果没找到再考虑右兄弟结点
//route存储方向，hashroute记录哈希值，注意往左走时应记录右兄弟节点哈希值，反之亦然（想想验证过程）
//t代表temporary
func (mn *MerkleNode) Find(data []byte, route []int, hashroute [][]byte) (bool, []int, [][]byte) {
	findFlag := false
	if bytes.Equal(mn.Data, data) {
		findFlag = true
		return findFlag, route, hashroute
	} else {
		if mn.LeftNode != nil {
			route_t := append(route, 0)
			hashroute_t := append(hashroute, mn.RightNode.Data)
			findFlag, route_t, hashroute_t = mn.LeftNode.Find(data, route_t, hashroute_t)
			if findFlag {
				return findFlag, route_t, hashroute_t
			} else {
				if mn.RightNode != nil {
					route_t = append(route, 1)
					hashroute_t = append(hashroute, mn.LeftNode.Data)
					findFlag, route_t, hashroute_t = mn.RightNode.Find(data, route_t, hashroute_t)
					if findFlag {
						return findFlag, route_t, hashroute_t //找到则返回_t，因为findFlag是1，所以会一直递归返回直到退出整个函数
					} else {
						return findFlag, route, hashroute //没找到返回原版（不返回_t是因为交易不在右子树，记录没有意义；因为findFlag是0，会继续查找）
					}
				}
			}
		} else {
			return findFlag, route, hashroute
		}
	}
	return findFlag, route, hashroute
}

// BackValidationRoute 调用Find函数判断对应txid是否存在在树中，并返回遍历的路径及路径上的哈希值
func (mt *MerkleTree) BackValidationRoute(txid []byte) ([]int, [][]byte, bool) {
	ok, route, hashroute := mt.RootNode.Find(txid, []int{}, [][]byte{})
	return route, hashroute, ok
}

// SPV 快速交易验证，route记录了路径方向，0代表左，1代表右，由此可以分辨出tempHash应当在左或右（如果顺序错了哈希值肯定对不上）
func SPV(txid, mtroothash []byte, route []int, hashroute [][]byte) bool {
	routeLen := len(route)
	var tempHash []byte
	tempHash = txid
	for i := routeLen - 1; i >= 0; i-- {
		if route[i] == 0 {
			catenateHash := append(tempHash, hashroute[i]...)
			hash := sha256.Sum256(catenateHash)
			tempHash = hash[:]
		} else if route[i] == 1 {
			catenateHash := append(tempHash, hashroute[0]...)
			hash := sha256.Sum256(catenateHash)
			tempHash = hash[:]
		} else {
			utils.Handle(errors.New("transaction not in block"))
		}
	}
	return bytes.Equal(tempHash, mtroothash)
}
