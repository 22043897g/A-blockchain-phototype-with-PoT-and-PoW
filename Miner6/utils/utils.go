package utils

import (
	"BlockchainInGo/constcoe"
	"bufio"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"github.com/mr-tron/base58"
	"golang.org/x/crypto/ripemd160"
	"io/ioutil"
	"log"
	"math"
	"math/big"
	"os"
	"strconv"
	"strings"
)

// Handle 错误处理
func Handle(err error) {
	if err != nil {
		log.Panic(err)
	}
}

// WriteBlockTime 记录出块时间
func WriteBlockTime(content string) {
	fd, err := os.OpenFile("time.txt", os.O_APPEND|os.O_CREATE, 0766)
	defer fd.Close()
	if err != nil {
		Handle(err)
	}
	w := bufio.NewWriter(fd)
	_, err2 := w.WriteString(content)
	if err2 != nil {
		Handle(err2)
	}
	w.Flush()
	fd.Sync()
}

// ReadBlockTime 读取出块时间
func ReadBlockTime(filename string) []int64 {
	var times []int64
	lines, err := ioutil.ReadFile(filename)
	if err != nil {
		Handle(err)
	} else {
		contents := string(lines)
		lines := strings.Split(contents, "\n")
		for _, line := range lines {
			if line != "" {
				tm, _ := strconv.ParseInt(line, 10, 64)
				times = append(times, tm)
			}
		}
	}
	return times
}

// AverageInterval 计算平均出块时间
func AverageInterval(times []int64) int64 {
	var Intervals []int64
	var sum int64
	lenTimes := len(times)
	temp := int(math.Min(10, float64(lenTimes)))
	j := lenTimes - 1
	//if temp == 1 {
	//	return times[0]
	//}
	for i := temp - 1; i > 0; i-- {
		Intervals = append(Intervals, times[j]-times[j-1])
		j--
	}
	for i := temp - 2; i >= 0; i-- {
		sum += Intervals[i]
	}
	average := sum / 9
	fmt.Println("Interval sum:", sum, "average:", average)
	return average
}

// Int64ToBytes 和 BytesToInt64 格式转换
func Int64ToBytes(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	Handle(err)
	return buff.Bytes()
}

func BytesToInt64(buf []byte) int64 {
	return int64(binary.BigEndian.Uint32(buf))
}

// FileExists 判断文件是否存在
func FileExists(fileAddr string) bool {
	if _, err := os.Stat(fileAddr); os.IsNotExist(err) {
		return false
	}
	return true
}

// PublicKeyHash 将公钥转换为公钥哈希
func PublicKeyHash(publicKey []byte) []byte {
	// 在比特币中是先用sha256再用ripemd160
	hashedPublicKey := sha256.Sum256(publicKey)
	hasher := ripemd160.New()
	_, err := hasher.Write(hashedPublicKey[:])
	Handle(err)
	PublicRipeMd := hasher.Sum(nil) //转为byte slice
	return PublicRipeMd
}

// CheckSum 将上个函数中的结果哈希两次，并取最后四位作为检查位
func CheckSum(ripeMdHash []byte) []byte {
	firstHash := sha256.Sum256(ripeMdHash)
	secondHash := sha256.Sum256(firstHash[:])
	return secondHash[:constcoe.ChecksumLength]
}

// Base58Encode 上面两个函数会生成256位地址，但比特币用的Base58
func Base58Encode(input []byte) []byte {
	encode := base58.Encode(input)
	return []byte(encode)
}

func Base58Decode(input []byte) []byte {
	decode, err := base58.Decode(string(input[:]))
	Handle(err)
	return decode
}

// PubHash2Address 从公钥哈希生成地址
func PubHash2Address(pubKeyHash []byte) []byte {
	networkVersionedHash := append([]byte{constcoe.NetworkVersion}, pubKeyHash...)
	checkSum := CheckSum(networkVersionedHash)
	finalHash := append(networkVersionedHash, checkSum...)
	address := Base58Encode(finalHash)
	return address
}

// Address2PubHash 地址转公钥哈希
func Address2PubHash(address []byte) []byte {
	pubKeyHash := Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-constcoe.ChecksumLength]
	return pubKeyHash
}

// Sign 使用私钥进行签名，r是随机大数，s是根据r、私钥和message生成的签名
func Sign(msg []byte, privKey ecdsa.PrivateKey) []byte {
	r, s, err := ecdsa.Sign(rand.Reader, &privKey, msg)
	Handle(err)
	signature := append(r.Bytes(), s.Bytes()...)
	return signature
}

// Verify 验证签名
func VerifySig(msg []byte, pubkey []byte, signature []byte) bool {
	curve := elliptic.P256()
	r := big.Int{}
	s := big.Int{}
	sigLen := len(signature)
	r.SetBytes(signature[:(sigLen / 2)])
	s.SetBytes(signature[(sigLen / 2):])

	x := big.Int{}
	y := big.Int{}
	keyLen := len(pubkey)
	x.SetBytes(pubkey[:(keyLen / 2)])
	y.SetBytes(pubkey[(keyLen / 2):])

	rawPubkey := ecdsa.PublicKey{curve, &x, &y}
	return ecdsa.Verify(&rawPubkey, msg, &r, &s)
}
