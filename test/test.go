package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"runtime"
	"time"
	"unsafe"

	btcec "github.com/cuckoobtcec"
)

var ecctimeall time.Duration = 0
var eccenall time.Duration = 0
var eccdeall time.Duration = 0
var ecchaddall time.Duration = 0
var ecchaddfieldall time.Duration = 0
var ecchmulall time.Duration = 0

var c = btcec.S256()

func testparbtcec(messages [btcec.TestNum]*big.Int) {
	privKey, _ := btcec.NewPrivateKey(c)
	pubKey := privKey.PubKey()
	var ciphers [btcec.TestNum]*btcec.Cipher
	runtime.GOMAXPROCS(btcec.Threadnum)
	for i := 0; i < btcec.TestNum; i++ {
		start1 := time.Now()
		ciphers[i] = btcec.Encrypt(pubKey, messages[i])
		cost1 := time.Since(start1)
		eccenall = eccenall + cost1
	}
	start2 := time.Now()
	for i := 0; i < btcec.TestNum; i++ {
		btcec.WG.Add(1)
		go btcec.ParDecrypt(privKey, ciphers[i])
	}
	btcec.WG.Wait()
	cost2 := time.Since(start2)
	eccdeall = eccenall + cost2
	for i := 0; i < btcec.TestNum; i++ {
		cipher1 := btcec.Encrypt(pubKey, messages[btcec.TestNum-1-i])
		start4 := time.Now()
		btcec.HomoAdd(ciphers[i], cipher1)
		cost4 := time.Since(start4)
		start5 := time.Now()
		btcec.HomoAddField(ciphers[i], cipher1)
		cost5 := time.Since(start5)
		start6 := time.Now()
		btcec.HomoMul(ciphers[i], messages[btcec.TestNum-1-i])
		cost6 := time.Since(start6)
		ecchaddall = ecchaddall + cost4
		ecchaddfieldall = ecchaddfieldall + cost5
		ecchmulall = ecchmulall + cost6
	}
}

func main() {

	path := "../genlist/T1.bin"
	isexist, _ := btcec.PathExists(path)
	if isexist == true {
		testnum := btcec.TestNum
		var msgmax = big.NewInt(int64(btcec.Mmax))
		var messages [btcec.TestNum]*big.Int
		for i := 0; i < testnum; i++ {
			msg, _ := rand.Int(rand.Reader, msgmax)
			if i%2 == 0 {
				messages[i] = msg
			} else {
				messages[i] = msg.Neg(msg)
			}
		}
		testparbtcec(messages)
		fmt.Printf("l1 = %d,l2 = %d\n", btcec.Jlen+1, btcec.Ilen+1)
		fmt.Printf("Exp-ElGamal encrypto %d times average cost=[%s]\n", testnum, eccenall/time.Duration(testnum))
		fmt.Printf("Exp-ElGamal (FastECDLP, Threadnum = %d) decrypto %d times average cost=[%s]\n", btcec.Threadnum, testnum, eccdeall/time.Duration(testnum))
		fmt.Printf("Exp-ElGamal h-add %d times average cost=[%s]\n", testnum, ecchaddall/time.Duration(testnum))
		fmt.Printf("Exp-ElGamal h-fieldadd %d times average cost=[%s]\n", testnum, ecchaddfieldall/time.Duration(testnum))
		fmt.Printf("Exp-ElGamal h-mul %d times average cost=[%s]\n", testnum, ecchmulall/time.Duration(testnum))
		fmt.Printf("Memory Space T1+T2: %f GB\n", float64(btcec.SizeOf(btcec.T1))/1024/1024/1024+float64(unsafe.Sizeof(btcec.T2x))/1024/1024/1024*2)
	} else {
		fmt.Println("Please run “go run genlist.go” to generate T1")
	}
}
