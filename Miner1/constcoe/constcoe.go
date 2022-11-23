package constcoe

const (
	InitDifficulty      = 21
	Address             = "0.0.0.0:8001"
	Refname             = "A"
	InitCoin            = 30 // 创建区块时的奖励
	TransactionPoolFile = "./tmp/transaction_pool.data"
	BCPath              = "./tmp/blocks"
	BCFile              = "./tmp/blocks/MANIFEST"
	ChecksumLength      = 4          // 用于生成钱包地址
	NetworkVersion      = byte(0x00) // 用于生成钱包地址
	Wallets             = "D:/tmp/wallets/"
	WalletsRefList      = "D:/tmp/ref_list/"
)
