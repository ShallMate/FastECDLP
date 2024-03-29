package btcec

import (
	"bufio"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"math/big"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	Ilen      = 10   // l2-1
	Imax      = 1024 // 1<< Ilen
	Treelen   = 2048 // imax*2
	TestNum   = 1000
	Jlen      = 24       // l1-1
	Jmax      = 16777216 // 1<<Jlen
	cuckoolen = 21810381 //jmax*1.3
)

var Threadnum = 8

// var Routinenum = 8
var WG sync.WaitGroup
var IsParTree = 0
var IsNormal = 0

var Mmax = Imax * Jmax * 2
var mflag = int64(Mmax)
var c = S256()

var T1 = Op_NewCuckoo()

var T2x [Imax]FieldVal
var T2y [Imax]FieldVal

var MapT1 map[uint64]uint32

var zero = big.NewInt(0)
var one = big.NewInt(1)
var three = big.NewInt(3)
var seven = big.NewInt(7)

type Point struct {
	X *big.Int
	Y *big.Int
}

type Cipher struct {
	c1x *big.Int
	c1y *big.Int
	c2x *big.Int
	c2y *big.Int
}

type FieldCipher struct {
	c1x *FieldVal
	c1y *FieldVal
	c1z *FieldVal
	c2x *FieldVal
	c2y *FieldVal
	c2z *FieldVal
}

func BuildTree(zs []*FieldVal) [Treelen]*FieldVal {
	var ZTree [Treelen]*FieldVal
	for i := 0; i < Imax; i++ {
		ZTree[i] = zs[i]
	}
	offset := Imax
	treelen := Imax*2 - 3
	//treelen1 := treelen - 1
	for i := 0; i < treelen; i += 2 {
		z := new(FieldVal)
		zmult := z.Mul2(ZTree[i], ZTree[i+1])
		//zmult.Normalize()
		ZTree[offset] = zmult
		offset = offset + 1
	}
	return ZTree
}

func GetInvTree(rootinv *FieldVal, ZTree [Treelen]*FieldVal) [Treelen]*FieldVal {
	var ZinvTree [Treelen]*FieldVal
	treelen := Imax*2 - 2
	prevfloorflag := treelen
	prevfloornum := 1
	thisfloorflag := treelen
	treeroot_inv := new(FieldVal)
	treeroot_inv.Set(rootinv)
	ZinvTree[prevfloorflag] = treeroot_inv
	for i := 0; i < Ilen; i++ {
		thisfloornum := prevfloornum * 2
		thisfloorflag = prevfloorflag - thisfloornum
		for f := 0; f < thisfloornum; f++ {
			thisindex := f + thisfloorflag
			ztreeindex := thisindex ^ 1
			z := new(FieldVal)
			ZinvTree[thisindex] = z.Mul2(ZTree[ztreeindex], ZinvTree[prevfloorflag+(f/2)])
		}
		prevfloorflag = thisfloorflag
		prevfloornum = prevfloornum * 2
	}
	return ZinvTree
}

func Encrypt(pubkey *PublicKey, m *big.Int) *Cipher {
	start1 := time.Now()
	r, _ := NewPrivateKey(c)
	rpkx, rpky := c.ScalarMult(pubkey.X, pubkey.Y, r.D.Bytes())
	cost1 := time.Since(start1)
	//fmt.Printf("btcecc encrypt cost=[%s]\n", cost1)
	GetEnc = GetEnc + cost1
	mGx, mGy := c.ScalarBaseMult(m.Bytes())
	//fmt.Println(mGx)
	if m.Cmp(zero) == -1 {
		mGy = mGy.Sub(c.P, mGy)
	}
	c2x, c2y := c.Add(mGx, mGy, rpkx, rpky)
	return &Cipher{r.PublicKey.X, r.PublicKey.Y, c2x, c2y}
}

func NormalEnc(pubkey *PublicKey, m *big.Int) *Cipher {
	start1 := time.Now()
	r, _ := NewPrivateKey(c)
	rpkx, rpky := c.ScalarMult(pubkey.X, pubkey.Y, r.D.Bytes())
	cost1 := time.Since(start1)
	//fmt.Printf("btcecc encrypt cost=[%s]\n", cost1)
	GetEnc = GetEnc + cost1
	mGx, mGy := c.ScalarBaseMult(m.Bytes())
	c2x, c2y := c.Add(mGx, mGy, rpkx, rpky)
	return &Cipher{r.PublicKey.X, r.PublicKey.Y, c2x, c2y}
}

func EncryptJob(pubkey *PublicKey, m *big.Int) *FieldCipher {
	r, _ := NewPrivateKey(c)
	rpkx, rpky := c.ScalarMult(pubkey.X, pubkey.Y, r.D.Bytes())
	mGx, mGy := c.ScalarBaseMult(m.Bytes())
	if m.Cmp(zero) == -1 {
		mGy = mGy.Sub(c.P, mGy)
	}
	c2x, c2y, c2z := c.Add1(mGx, mGy, rpkx, rpky)
	c1x, c1y := c.bigAffineToField(r.PublicKey.X, r.PublicKey.Y)
	c1z := new(FieldVal).SetInt(1)
	return &FieldCipher{c1x, c1y, c1z, c2x, c2y, c2z}
}

var GetEnc time.Duration = 0
var GetmG time.Duration = 0
var GetX21 time.Duration = 0
var GetTree1 time.Duration = 0
var GetInv time.Duration = 0
var GetTree2 time.Duration = 0
var BSGS time.Duration = 0
var Verify time.Duration = 0
var GetHash time.Duration = 0
var GetX3 time.Duration = 0
var GetSearch time.Duration = 0

func BytesToUint32(b []byte) Hash {
	_ = b[3]
	return Hash(b[3]) | Hash(b[2])<<8 | Hash(b[1])<<16 | Hash(b[0])<<24
}

func GetM(mGx *big.Int, ZinvTree [Treelen]*FieldVal, p *FieldVal, fmGx, fmGy *FieldVal, t, start, end, jmax int, m_new []int64, m_bool []bool, overthreadnum chan int) {
	//time.Sleep(time.Duration(4) * time.Second)
	for j := start; j < end; j++ {
		if j == 0 {
			leftxbytes := mGx.Bytes()
			i, ok := T1.Op_search(leftxbytes)
			if ok {
				m := int64(i)
				TestmGx, _ := c.ScalarBaseMult(big.NewInt(m).Bytes())
				r1 := mGx.Cmp(TestmGx)
				if r1 == 0 {
					m_bool[t] = true
					m_new[t] = m
					break
				}
			}
		}

		ft2x, ft2y := T2x[j], T2y[j]
		/*
			ft2x := T2x[j]
			t2x := new(big.Int).SetBytes(ft2x.Bytes()[:])
			t2y := GetY(t2x)
			ft2y := new(FieldVal).SetByteSlice(t2y.Bytes())
		*/

		leftx, invleftx := c.NewGetx3(fmGx, fmGy, &ft2x, &ft2y, ZinvTree[j], p)
		leftxbytes := leftx.Bytes()

		i, ok := T1.Op_search(leftxbytes)
		if ok {
			m1 := int64(j+1) * int64(jmax) * 2
			m2 := int64(i)
			m := m1 + m2
			TestmGx, _ := c.ScalarBaseMult(big.NewInt(m).Bytes())
			r1 := mGx.Cmp(TestmGx)
			if r1 == 0 {
				m_bool[t] = true
				m_new[t] = m
				break
			}

			m = m1 - m2
			TestmGx, _ = c.ScalarBaseMult(big.NewInt(m).Bytes())
			r1 = mGx.Cmp(TestmGx)
			if r1 == 0 {
				m_bool[t] = true
				m_new[t] = m
				break
			}
		}
		leftxbytes = invleftx.Bytes()
		i, ok = T1.Op_search(leftxbytes)
		if ok {
			m1 := int64(j+1) * int64(jmax) * 2
			m2 := int64(i)
			m := -(m1 + m2)
			TestmGx, _ := c.ScalarBaseMult(big.NewInt(m).Bytes())
			r1 := mGx.Cmp(TestmGx)
			if r1 == 0 {
				m_bool[t] = true
				m_new[t] = m
				break
			}

			m = -(m1 - m2)
			TestmGx, _ = c.ScalarBaseMult(big.NewInt(m).Bytes())
			r1 = mGx.Cmp(TestmGx)
			if r1 == 0 {
				m_bool[t] = true
				m_new[t] = m
				break
			}
		}
	}
	overthreadnum <- 1
}

func ParDecrypt(priv *PrivateKey, cipher *Cipher) (*big.Int, string) {
	//start1 := time.Now()
	var m int64 = mflag
	skc1x, skc1y := c.ScalarMult(cipher.c1x, cipher.c1y, priv.D.Bytes())
	if skc1x.Cmp(cipher.c2x) == 0 {
		return zero, ""
	}
	//ZTree := make([]*FieldVal, Imax*2, Imax*2)
	//ZinvTree := make([]*FieldVal, Imax*2, Imax*2)
	//var ZTree [Treelen]*FieldVal
	//var ZinvTree [Treelen]*FieldVal
	inv_skc1y := new(big.Int)
	inv_skc1y.Add(c.P, inv_skc1y)
	inv_skc1y.Sub(inv_skc1y, skc1y)
	mGx, mGy := c.Add(cipher.c2x, cipher.c2y, skc1x, inv_skc1y)
	//fmt.Println(mGx)
	fmGx, fmGy := c.bigAffineToField(mGx, mGy)
	zs := make([]*FieldVal, Imax)
	for i := 0; i < Imax; i++ {
		ft2x := T2x[i]
		zs[i] = c.Getz3(fmGx, &ft2x)
		zs[i].Normalize()
		if zs[i].Equals(fieldZero) == true {
			m = int64(Jmax*2) * int64(i+1)
			mbigint := big.NewInt(m)
			_, TestmGy := c.ScalarBaseMult(mbigint.Bytes())
			r1 := mGx.Cmp(TestmGy)
			if r1 == 0 {
				WG.Done()
				return big.NewInt(m), "secuess"
			}
			WG.Done()
			return new(big.Int).Neg(mbigint), "secuess"
		}
	}
	//runtime.GOMAXPROCS(runtime.NumCPU())
	runtime.GOMAXPROCS(Threadnum)
	ZTree := BuildTree(zs)
	treeroot_inv := new(FieldVal).Set(ZTree[Treelen-2]).Inverse()
	ZinvTree := GetInvTree(treeroot_inv, ZTree)
	p := new(FieldVal).SetByteSlice(c.P.Bytes())

	overthreadnum := make(chan int, Threadnum)
	batch := Imax / (Threadnum)
	m_new := make([]int64, Threadnum)
	m_bool := make([]bool, Threadnum)
	for t := 0; t < Threadnum; t++ {
		m_new[t] = m
		m_bool[t] = false
	}

	for t := 0; t < Threadnum; t++ {
		go GetM(mGx, ZinvTree, p, fmGx, fmGy, t, t*batch, (t+1)*batch, Jmax, m_new, m_bool, overthreadnum)
	}

	for i := 0; i < Threadnum; i++ {
		<-overthreadnum
	}
	acc_c := 0
	for t := 0; t < Threadnum; t++ {
		if m_bool[t] {
			acc_c += 1
			m = m_new[t]
			TestmGx, _ := c.ScalarBaseMult(big.NewInt(m).Bytes())
			r1 := mGx.Cmp(TestmGx)
			if r1 == 0 {
				if acc_c > 1 {
					fmt.Println("getM 多次", acc_c)
				}
				WG.Done()
				return big.NewInt(m), "sucess"
			}

		}
	}
	WG.Done()
	fmt.Println("解密失败", acc_c)
	return big.NewInt(0), "decrypt error 2"
}

func GetY(t2x *big.Int) *big.Int {
	t2y2 := new(big.Int).Exp(t2x, three, c.P)
	t2y2 = t2y2.Add(t2y2, seven)
	t2y2 = t2y2.Mod(t2y2, c.P)
	t2y := t2y2.Sqrt(t2y2)
	t2y = t2y.Mod(t2y, c.P)
	inv_t2y := t2y.Sub(c.P, t2y)
	return inv_t2y
}

func NormalDecrypt(priv *PrivateKey, cipher *Cipher) (*big.Int, string) {
	var m int64 = mflag
	skc1x, skc1y := c.ScalarMult(cipher.c1x, cipher.c1y, priv.D.Bytes())
	if skc1x.Cmp(cipher.c2x) == 0 {
		return zero, ""
	}
	inv_skc1y := new(big.Int)
	inv_skc1y.Add(c.P, inv_skc1y)
	inv_skc1y.Sub(inv_skc1y, skc1y)
	mGx, mGy := c.Add(cipher.c2x, cipher.c2y, skc1x, inv_skc1y)
	start := time.Now()
	for j := 0; j < Imax; j++ {
		if j == 0 {
			// hash time
			leftxbytes := mGx.Bytes()[:8]
			x64 := binary.BigEndian.Uint64(leftxbytes)
			i, ok := MapT1[x64]
			if ok {
				m = int64(i)
				cost := time.Since(start)
				GetmG = GetmG + cost
				break
			}
		}
		//ft2x, ft2y := T2x[j], T2y[j]
		ft2x := T2x[j]
		t2x := new(big.Int).SetBytes(ft2x.Bytes()[:])
		t2y := GetY(t2x)
		//ft2y := new(big.Int).SetBytes(ft2y.Bytes()[:])
		leftx, _ := c.Add(mGx, mGy, t2x, t2y)
		//leftx := c.Getx3(fmGx, fmGy, ft2x, ft2y, ZinvTree[j])
		leftxbytes := leftx.Bytes()[:8]
		x64 := binary.BigEndian.Uint64(leftxbytes)
		i, ok := MapT1[x64]
		if ok {
			m = int64(j)*int64(Jmax) + int64(i)
			cost := time.Since(start)
			GetmG = GetmG + cost
			break
		}
	}
	return big.NewInt(m), "sucess"
}

func GetZS(c *KoblitzCurve, zs []*FieldVal, fmGx *FieldVal, start int, end int) {
	for i := start; i < end; i++ {
		ft2x := T2x[i]
		zs[i] = c.Getz3(fmGx, &ft2x)
	}
}

func HomoAddField(c1 *Cipher, c2 *Cipher) *FieldCipher {
	c1x, c1y, c1z := c.Add1(c1.c1x, c1.c1y, c2.c1x, c2.c1y)
	c2x, c2y, c2z := c.Add1(c1.c2x, c1.c2y, c2.c2x, c2.c2y)
	return &FieldCipher{c1x, c1y, c1z, c2x, c2y, c2z}
}

func HomoAddField1(c1 *FieldCipher, c2 *FieldCipher) *FieldCipher {
	c1x, c1y, c1z := new(FieldVal), new(FieldVal), new(FieldVal)
	c2x, c2y, c2z := new(FieldVal), new(FieldVal), new(FieldVal)
	c.AddGeneric(c1.c1x, c1.c1y, c1.c1z, c2.c1x, c2.c1y, c2.c1z, c1x, c1y, c1z)
	c.AddGeneric(c1.c2x, c1.c2y, c1.c2z, c2.c2x, c2.c2y, c2.c2z, c2x, c2y, c2z)
	return &FieldCipher{c1x, c1y, c1z, c2x, c2y, c2z}
}

func HomoAdd(c1 *Cipher, c2 *Cipher) *Cipher {
	c1x, c1y := c.Add(c1.c1x, c1.c1y, c2.c1x, c2.c1y)
	c2x, c2y := c.Add(c1.c2x, c1.c2y, c2.c2x, c2.c2y)
	return &Cipher{c1x, c1y, c2x, c2y}
}

func HomoAddPlainText(c1 *Cipher, c2 *big.Int) *Cipher {
	c2x, c2y := c.ScalarBaseMult(c2.Bytes())
	c2x, c2y = c.Add(c1.c2x, c1.c2y, c2x, c2y)
	return &Cipher{c1.c1x, c1.c1y, c2x, c2y}
}

func HomoMul(c1 *Cipher, k *big.Int) *Cipher {
	c1x, c1y := c.ScalarMult(c1.c1x, c1.c1y, k.Bytes())
	c2x, c2y := c.ScalarMult(c1.c2x, c1.c2y, k.Bytes())
	return &Cipher{c1x, c1y, c2x, c2y}
}

func HomoMulField(c1 *Cipher, k *big.Int) *FieldCipher {
	c1x, c1y, c1z := c.ScalarMultField(c1.c1x, c1.c1y, k.Bytes())
	c2x, c2y, c2z := c.ScalarMultField(c1.c2x, c1.c2y, k.Bytes())
	return &FieldCipher{c1x, c1y, c1z, c2x, c2y, c2z}
}

func ConvertCipher(fieldc *FieldCipher) *Cipher {
	c1x, c1y := c.fieldJacobianToBigAffine(fieldc.c1x, fieldc.c1y, fieldc.c1z)
	c2x, c2y := c.fieldJacobianToBigAffine(fieldc.c2x, fieldc.c2y, fieldc.c2z)
	return &Cipher{c1x, c1y, c2x, c2y}
}

func ReadT1() {
	T1_file, _ := os.Open("/home/lgw/go/src/github.com/cuckoobtcec/genlist/Tx24.bin")
	T1_dec := gob.NewDecoder(T1_file)
	T1_dec.Decode(T1)
	T1_file.Close()
}

func ReadT2() {
	var j int64 = 1
	t1lastx, t1lasty := c.ScalarMult(c.Gx, c.Gy, big.NewInt(int64(Jmax)).Bytes())
	t2x, t2y := c.ScalarMult(t1lastx, t1lasty, zero.Bytes())
	for ; j < int64(Imax); j++ {
		if j >= 1 {
			t2x, t2y = c.Add(t2x, t2y, t1lastx, t1lasty)
		}
		inv_t2y := new(big.Int)
		inv_t2y.Add(c.P, inv_t2y)
		inv_t2y.Sub(inv_t2y, t2y)
		ft2x, ft2y := c.bigAffineToField(t2x, inv_t2y)
		T2x[j] = *ft2x
		T2y[j] = *ft2y
	}
}

func ReadT1AsMap() {
	var i int64 = 1
	filename := "/home/lgw/go/src/github.com/cuckoobtcec/genlist/Tx28.txt"
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	x := big.NewInt(0)
	MapT1 = make(map[uint64]uint32)
	rd := bufio.NewReader(file)
	for {
		line, err := rd.ReadString('\n')
		if err != nil || io.EOF == err {
			break
		} else {
			line = strings.Replace(line, "\n", "", -1)
			x, _ = new(big.Int).SetString(line, 10)
			MapT1[x.Uint64()] = uint32(i)
			if i == int64(Jmax) {
				file.Close()
				break
			}
			i++
		}
	}
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func ReadT1frombin() {
	path := "../genlist/T1.bin"
	isexist, _ := PathExists(path)
	if isexist == true {
		T1.Load(path)
	}
}

func init() {
	if IsNormal == 1 {
		ReadT1AsMap()
	} else {
		ReadT1frombin()
	}
	var j int64 = 0
	t1lastx, t1lasty := c.ScalarMult(c.Gx, c.Gy, big.NewInt(int64(Jmax*2)).Bytes())
	t2x, t2y := c.ScalarMult(t1lastx, t1lasty, one.Bytes())
	for ; j < int64(Imax); j++ {
		//fmt.Printf("%d\n", j)
		if j >= 1 {
			t2x, t2y = c.Add(t2x, t2y, t1lastx, t1lasty)
		}
		inv_t2y := new(big.Int)
		inv_t2y.Add(c.P, inv_t2y)
		inv_t2y.Sub(inv_t2y, t2y)
		ft2x, ft2y := c.bigAffineToField(t2x, inv_t2y)
		T2x[j] = *ft2x
		T2y[j] = *ft2y
	}

}
