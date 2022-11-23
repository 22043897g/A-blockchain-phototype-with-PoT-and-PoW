package wallet

import (
	"BlockchainInGo/constcoe"
	"BlockchainInGo/utils"
	"bytes"
	"encoding/gob"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type RefList map[string]string //钱包地址到钱包别名的映射

func (r *RefList) Save() {
	filename := constcoe.WalletsRefList + "ref_list.data"
	var content bytes.Buffer
	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(r)
	utils.Handle(err)
	err = ioutil.WriteFile(filename, content.Bytes(), 0644)
	utils.Handle(err)
}

// Update 扫描电脑中所有.wlt文件
func (r *RefList) Update() {
	err := filepath.Walk(constcoe.Wallets, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		fileName := f.Name() //只是文件名称，不带路径
		if strings.Compare(fileName[len(fileName)-4:], ".wlt") == 0 {
			//如果钱包列表已有该钱包地址对应的映射则不进行操作
			_, ok := (*r)[fileName[:len(fileName)-4]]
			if !ok {  //没有映射则添加映射，且钱包别名为空
				(*r)[fileName[:len(fileName)-4]] = ""
			}
		}
		return nil
	})
	utils.Handle(err)
}

// LoadRefList 从本地读取RefList，如果没有则创建新的并加载所有wlt文件
func LoadRefList() *RefList {
	filename := constcoe.WalletsRefList + "ref_list.data"
	var reflist RefList
	if utils.FileExists(filename) {
		fileContent, err := ioutil.ReadFile(filename)
		utils.Handle(err)
		decoder := gob.NewDecoder(bytes.NewBuffer(fileContent))
		err = decoder.Decode(&reflist)
		utils.Handle(err)
	} else {
		reflist = make(RefList)
		reflist.Update()
	}
	return &reflist
}

// BindRef 为每个地址绑定别称（为了方便展示，实际比特币并没有）
func (r *RefList) BindRef(address, refname string) {
	(*r)[address] = refname
}

// FindRef 通过别名找到钱包地址
func (r *RefList) FindRef(refname string) (string, error) {
	temp := ""
	for key, val := range *r {
		if val == refname {
			temp = key
			break
		}
	}
	if temp == "" {
		err := errors.New("The refname is not found")
		return temp, err
	}
	return temp, nil
}
