package test

func Tx1() {
	SendRefName("A", "B", 25, 2)
}

func Tx2() {
	SendRefName("B", "C", 10, 2)
}

func Tx3() {
	SendRefName("C", "B", 3, 2)
}
