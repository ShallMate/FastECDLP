package main

import (
	"fmt"
	"math/big"
	"runtime"

	btcec "github.com/cuckoobtcec"
)

var XS [btcec.Jmax][]byte

func Getxbytes(c *btcec.KoblitzCurve, i int64) {
	x, _ := c.ScalarBaseMult(big.NewInt(i).Bytes())
	XS[i-1] = x.Bytes()
}

func main() {
	maxexp := 1 << btcec.Jlen
	cuckoo := btcec.Op_NewCuckoo()
	//fmt.Println(cuckoo)
	runtime.GOMAXPROCS(btcec.Threadnum)
	c := btcec.S256()
	var i int64 = 1
	for ; i <= int64(maxexp); i++ {
		go Getxbytes(c, i)
	}
	fmt.Println("Generate XS")
	cuckoo.Op_insert(XS)
	cuckoo.Save("./T1.bin")
}
