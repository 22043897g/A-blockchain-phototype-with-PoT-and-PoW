package addresses

import (
	"BlockchainInGo/constcoe"
	"fmt"
	"github.com/go-redis/redis"
	"math/rand"
	"strconv"
)

var rdb *redis.Client

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

func SaveNewAddress(address string) {
	RedisInit()
	//因为地址数组下标从0开始，所以总数正好是存储新地址的索引值
	res, _ := rdb.Get("AddressAmount").Result()
	key := "A" + res
	println(key)
	rdb.Set(key, address, 0)
	//地址总数+1后存回
	num, _ := strconv.ParseInt(res, 10, 64)
	newnum := num + 1
	value := strconv.FormatInt(newnum, 10)
	rdb.Set("AddressAmount", value, 0)
	println("New address:", address, " saved.")
}

func ReadAllAddress() []string {
	RedisInit()
	addrs := []string{}
	res, _ := rdb.Get("AddressAmount").Result()
	println(res)
	num, _ := strconv.ParseInt(res, 10, 64)
	for i := 0; i < int(num); i++ {
		key := "A" + strconv.FormatInt(int64(i), 10)
		addr, _ := rdb.Get(key).Result()
		addrs = append(addrs, addr)
	}
	return addrs
}

// 查询是否记录了某地址
func CheckAddress(address string) bool {
	RedisInit()
	res, _ := rdb.Get("AddressAmount").Result()
	num, _ := strconv.ParseInt(res, 10, 64)
	for i := 0; i < int(num); i++ {
		key := "A" + strconv.FormatInt(int64(i), 10)
		addr, _ := rdb.Get(key).Result()
		if addr == address {
			return true
		}
	}
	return false
}

// PortInit 初始化端口
func PortInit() {
	RedisInit()
	rdb.Set("AddressAmount", 0, 0)
	for i := 1; i <= 10; i++ {
		s := "0.0.0.0:800" + strconv.FormatInt(int64(i), 10)
		rdb.Set(s, "F", 0)
	}
	AddPort("0.0.0.0:8001")
	AddPort("0.0.0.0:8002")
	AddPort("0.0.0.0:8003")
}

func AddPort(addr string) {
	RedisInit()
	rdb.Set(addr, "T", 0)
	res, _ := rdb.Get("AddressAmount").Result()
	num, _ := strconv.ParseInt(res, 10, 64)
	newnum := num + 1
	value := strconv.FormatInt(newnum, 10)
	rdb.Set("AddressAmount", value, 0)
}

func DelPort(addr string) {
	RedisInit()
	rdb.Set(addr, "F", 0)
}

func ArrayShuffle(slice []string) {
	// 遍历循环打乱
	for len(slice) > 0 {
		n := len(slice)
		randIndex := rand.Intn(n)
		slice[n-1], slice[randIndex] = slice[randIndex], slice[n-1]
		slice = slice[:n-1]
	}
}

func AlivePort() []string {
	var AliveAddrs []string
	RedisInit()
	for i := 1; i <= 10; i++ {
		s := ""
		if i == 10 {
			s = "0.0.0.0:8010"
		} else {
			s = "0.0.0.0:800" + strconv.FormatInt(int64(i), 10)
		}
		flag, _ := rdb.Get(s).Result()
		if flag == "T" {
			AliveAddrs = append(AliveAddrs, s)
		}
	}
	//ArrayShuffle(AliveAddrs)
	for i := 0; i < len(AliveAddrs); i++ {
		if AliveAddrs[i] == constcoe.Address {
			AliveAddrs[i], AliveAddrs[len(AliveAddrs)-1] = AliveAddrs[len(AliveAddrs)-1], AliveAddrs[i]
		}
	}
	return AliveAddrs
}
