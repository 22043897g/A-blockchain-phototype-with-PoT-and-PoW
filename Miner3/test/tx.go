package test

func Tx1() {
	SendRefName("A", "B", 25, 10)
}

func Tx2() {
	SendRefName("B", "C", 10, 8)
}

func Tx3() {
	SendRefName("C", "B", 10, 5)
}
