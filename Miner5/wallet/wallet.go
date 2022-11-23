package wallet

import (
	"BlockchainInGo/constcoe"
	"BlockchainInGo/utils"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/gob"
	"errors"
	"fmt"
	"io/ioutil"
)

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

func NewKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()

	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	utils.Handle(err)
	publicKey := append(privateKey.PublicKey.X.Bytes(), privateKey.PublicKey.Y.Bytes()...) //公钥是个点
	return *privateKey, publicKey
}

func (w *Wallet) Address() []byte {
	pubHash := utils.PublicKeyHash(w.PublicKey)
	return utils.PubHash2Address(pubHash)
}

func NewWallet() *Wallet {
	privatekey, publickkey := NewKeyPair()
	wallet := Wallet{privatekey, publickkey}
	return &wallet
}

// Save 钱包保存到本地
func (w *Wallet) Save() {
	filename := constcoe.Wallets + string(w.Address()) + ".wlt"
	var content bytes.Buffer
	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(w)
	utils.Handle(err)
	err = ioutil.WriteFile(filename, content.Bytes(), 0644)
	utils.Handle(err)
}

// LoadWallet 读取钱包
func LoadWallet(address string) *Wallet {
	filename := constcoe.Wallets + address + ".wlt"
	if !utils.FileExists(filename) {
		utils.Handle(errors.New("wrong address"))
	}
	var w Wallet
	gob.Register(elliptic.P256())
	fileContent, err := ioutil.ReadFile(filename)
	utils.Handle(err)
	decoder := gob.NewDecoder(bytes.NewBuffer(fileContent))
	err = decoder.Decode(&w)
	utils.Handle(err)
	return &w
}

func CreateWallet(refname string) {
	newWallet := NewWallet()
	reflist := LoadRefList()
	address, _ := reflist.FindRef(refname)
	//同一名称不能有两个钱包
	if address != "" {
		println("This refname all ready have a wallet.")
		return
	}

	newWallet.Save()
	// 在钱包列表中加入新钱包并存储
	reflist.BindRef(string(newWallet.Address()), refname)
	reflist.Save()
	fmt.Println("Succeed in creating wallet.")
}
