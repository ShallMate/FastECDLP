package btcec

import (
	"encoding/gob"
	"fmt"
	"os"
)

const (
	EMPTY   uint8 = 0
	MAXITER int   = 500
)

type Key uint32

type Value uint32

type Op_Cuckoo struct {
	Table_v []Value
	Table_k []Key
	// Hash_index []uint8
	nentries int
	maxsize  Hash
}

func Op_NewCuckoo(logsize int) *Op_Cuckoo {
	nentries := (1 << uint(logsize))
	size := int(float64(nentries) * 1.3)
	c := &Op_Cuckoo{
		Table_v: make([]Value, size, size),
		Table_k: make([]Key, size, size),
		// Hash_index: make([]uint8, size),
		nentries: nentries,
		maxsize:  Hash(size),
	}
	return c
}

func (c *Op_Cuckoo) Op_search(X []byte) (v Value, ok bool) {

	for i := 0; i < 3; i++ {
		start := i * 8
		end := start + 4
		x := BytesToUint32(X[end : end+4])
		x_key := Key(x)
		h := BytesToUint32(X[start:end]) % c.maxsize
		if c.Table_k[h] == x_key {
			return c.Table_v[h], true
		}
	}
	return 0, false
}

func (c *Op_Cuckoo) Op_insert(data [][]byte) {
	Hash_index := make([]uint8, c.maxsize)

	for i := 0; i < c.nentries; i++ {
		v := Value(i + 1)
		old_hash_id := uint8(1)
		j := 0
		for ; j < MAXITER; j++ {
			X := data[v-1]
			start := (old_hash_id - 1) * 8
			end := start + 4
			x := BytesToUint32(X[end : end+4])
			x_key := Key(x)
			h := BytesToUint32(X[start:end]) % c.maxsize
			hash_id_address := &Hash_index[h]
			key_index_address := &c.Table_v[h]
			key_address := &c.Table_k[h]

			if *hash_id_address == EMPTY {
				*hash_id_address = old_hash_id
				*key_index_address = v
				*key_address = x_key
				break
			} else {
				v, *key_index_address = *key_index_address, v
				old_hash_id, *hash_id_address = *hash_id_address, old_hash_id
				x_key, *key_address = *key_address, x_key
				old_hash_id = old_hash_id%3 + 1
			}
		}
		if j == MAXITER-1 {
			fmt.Println("insert failed, ", i)
		}

	}
}

func (c *Op_Cuckoo) Save(save_file_path string) {
	// save
	fmt.Println("start save file T1")
	f, _ := os.Create(save_file_path)
	defer f.Close()
	enc := gob.NewEncoder(f)
	enc.Encode(c)
}

func (c *Op_Cuckoo) Load(save_file_path string) {
	T1_file, _ := os.Open(save_file_path)
	defer T1_file.Close()
	T1_dec := gob.NewDecoder(T1_file)
	T1_dec.Decode(c)
}
