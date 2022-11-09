package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	btcec "github.com/cuckoobtcec"
)

var ecctimeall time.Duration = 0
var eccenall time.Duration = 0
var eccdeall time.Duration = 0
var ecchaddall time.Duration = 0
var ecchaddfieldall time.Duration = 0
var ecchmulall time.Duration = 0

var c = btcec.S256()

func testparbtcec(messages []*big.Int) {
	privKey, _ := btcec.NewPrivateKey(c)
	pubKey := privKey.PubKey()
	err := 0
	for i := 0; i < btcec.TestNum; i++ {
		fmt.Println("hecc Eecryption message : ", messages[i])
		start1 := time.Now()
		cipher := btcec.Encrypt(pubKey, messages[i])
		cost1 := time.Since(start1)
		//fmt.Printf("btcecc encrypt cost=[%s]\n", cost1)
		eccenall = eccenall + cost1
		//plaintext := ""
		start2 := time.Now()
		plaintext, ok := btcec.ParDecrypt(privKey, cipher)
		cost2 := time.Since(start2)
		//fmt.Printf("btcecc decrypt cost=[%s]\n", cost2)
		eccdeall = eccdeall + cost2
		//cost3 := cost1 + cost2
		//fmt.Printf("btcecc all cost=[%s]\n", cost3)

		cipher1 := btcec.Encrypt(pubKey, messages[btcec.TestNum-1-i])
		start4 := time.Now()
		btcec.HomoAdd(cipher, cipher1)
		cost4 := time.Since(start4)
		start5 := time.Now()
		btcec.HomoAddField(cipher, cipher1)
		cost5 := time.Since(start5)
		start6 := time.Now()
		btcec.HomoMul(cipher, messages[btcec.TestNum-1-i])
		cost6 := time.Since(start6)

		ecchaddall = ecchaddall + cost4
		ecchaddfieldall = ecchaddfieldall + cost5
		ecchmulall = ecchmulall + cost6
		fmt.Println("hecc Decryption Result : ", plaintext)
		if ok != "sucess" {
			err = err + 1
			fmt.Println(messages[i])
		}
	}
	fmt.Println("The number of decryption fail:", err)
}

func main() {
	path := "../genlist/Tx24.bin"
	isexist, _ := btcec.PathExists(path)
	if isexist == true {

		testnum := btcec.TestNum
		var msgmax = big.NewInt(int64(btcec.Mmax))
		messages := make([]*big.Int, testnum)
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
		fmt.Printf("Exp-ElGamal decrypto %d times average cost=[%s]\n", testnum, eccdeall/time.Duration(testnum))
		fmt.Printf("Exp-ElGamal h-add %d times average cost=[%s]\n", testnum, ecchaddall/time.Duration(testnum))
		fmt.Printf("Exp-ElGamal h-fieldadd %d times average cost=[%s]\n", testnum, ecchaddfieldall/time.Duration(testnum))
		fmt.Printf("Exp-ElGamal h-mul %d times average cost=[%s]\n", testnum, ecchmulall/time.Duration(testnum))
		fmt.Printf("memory: %f GB\n", float64(btcec.SizeOf(btcec.T1))/1024/1024/1024+float64(btcec.SizeOf(btcec.T2x))/1024/1024/1024)
	} else {
		fmt.Println("Please run “go run genlist.go” to generate T1")
	}
}
