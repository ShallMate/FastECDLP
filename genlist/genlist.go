package main

import (
	"fmt"
	"math/big"

	btcec "github.com/cuckoobtcec"
)

func main() {
	maxexp := 1 << btcec.Jlen
	cuckoo := btcec.Op_NewCuckoo(btcec.Jlen)
	XS := make([][]byte, maxexp, maxexp)
	//fmt.Println(cuckoo)
	c := btcec.S256()
	var i int64 = 2
	//var k int64 = 1
	x, y := c.ScalarBaseMult(big.NewInt(1).Bytes())
	XS[0] = x.Bytes()
	for ; i <= int64(maxexp); i++ {
		//fmt.Printf("%d\n", i)
		x, y = c.Add(x, y, c.Gx, c.Gy)
		XS[i-1] = x.Bytes()
	}
	fmt.Println("Generate XS")
	cuckoo.Op_insert(XS)
	cuckoo.Save("./Tx24.bin")
}
