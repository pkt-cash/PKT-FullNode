package pcutil

import (
	"bytes"
	crand "crypto/rand"
	"testing"
	"time"

	b2b_dchest "github.com/dchest/blake2b"
	b2b_x "golang.org/x/crypto/blake2b"
)

const (
	Blake2bBenchmarkEnabled        = false //disable to remove noise from test
	Blake2bBenchmarkIterationCount = 10000000
)

// note: benchmark output as a test error
func TestBlake2bBenchmark_256(t *testing.T) {
	if !Blake2bBenchmarkEnabled { //filter from test results
		return
	}
	//create a random test hash (used by both)
	sampleHash32 := [32]byte{}
	crand.Read(sampleHash32[:])

	//perform x blake2b
	final_x := make([]byte, 32) //use a slice (match actual code)
	copy(final_x[:], sampleHash32[:])
	start := time.Now()
	for i := 0; i < Blake2bBenchmarkIterationCount; i++ {
		hash := b2b_x.Sum256(final_x[:])
		copy(final_x[:32], hash[:32]) //overwrite original
	}
	d_x := time.Since(start)

	//sanity check
	//t.Errorf("sample hash [%02X]", sampleHash32[:])
	//t.Errorf("x      hash [%02X]", final_x[:])

	//perform dchest blake2b
	final_dchest := make([]byte, 32) //use a slice (match current code)
	copy(final_dchest[:], sampleHash32[:])
	start = time.Now()
	for i := 0; i < Blake2bBenchmarkIterationCount; i++ {
		//copied from current code implementation
		b2 := b2b_dchest.New256()
		_, err := b2.Write(final_dchest[:])
		if err != nil {
			break
		}
		b2.Sum(final_dchest[:0]) //overwrite original
		//if we change dchest implementation (as below), still ~twice as slow
		//hash := b2b_dchest.Sum256(final_dchest)
		//copy(final_dchest[:32], hash[:32])
	}
	d_dchest := time.Since(start)

	//sanity check
	//t.Errorf("sample hash [%02X]", sampleHash32[:])
	//t.Errorf("dchest hash [%02X]", final_dchest[:])

	if !bytes.Equal(final_x, final_dchest) { //ensure final hashes match
		t.Errorf("%s - hash mismatch", t.Name())
	} else {
		//spit out benchmark result in place of test result
		t.Errorf("%s - %d iterations: x took %dms, dchest took %dms",
			t.Name(), Blake2bBenchmarkIterationCount,
			d_x.Milliseconds(), d_dchest.Milliseconds())
	}
}

// note: benchmark output as a test error
func TestBlake2bBenchmark_512(t *testing.T) {
	if !Blake2bBenchmarkEnabled { //filter from test results
		return
	}
	//create a random test hash (used by both)
	sampleHash64 := [64]byte{}
	crand.Read(sampleHash64[:])

	//perform x blake2b
	final_x := make([]byte, 64) //use a slice (match actual code)
	copy(final_x[:], sampleHash64[:])
	start := time.Now()
	for i := 0; i < Blake2bBenchmarkIterationCount; i++ {
		hash := b2b_x.Sum512(final_x[:])
		copy(final_x[:64], hash[:64]) //overwrite original
	}
	d_x := time.Since(start)

	//sanity check
	//t.Errorf("sample hash [%02X]", sampleHash64[:])
	//t.Errorf("x      hash [%02X]", final_x[:])

	//perform dchest blake2b
	final_dchest := make([]byte, 64) //use a slice (match current code)
	copy(final_dchest[:], sampleHash64[:])
	start = time.Now()
	for i := 0; i < Blake2bBenchmarkIterationCount; i++ {
		//copied from current code implementation
		b2 := b2b_dchest.New512()
		_, err := b2.Write(final_dchest[:])
		if err != nil {
			break
		}
		b2.Sum(final_dchest[:0]) //overwrite original
		//if we change dchest implementation (as below), still ~twice as slow
		//hash := b2b_dchest.Sum512(final_dchest)
		//copy(final_dchest[:64], hash[:64])
	}
	d_dchest := time.Since(start)

	//sanity check
	//t.Errorf("sample hash [%02X]", sampleHash64[:])
	//t.Errorf("dchest hash [%02X]", final_dchest[:])

	if !bytes.Equal(final_x, final_dchest) { //ensure final hashes match
		t.Errorf("%s - hash mismatch", t.Name())
	} else {
		//spit out benchmark result in place of test result
		t.Errorf("%s - %d iterations: x took %dms, dchest took %dms",
			t.Name(), Blake2bBenchmarkIterationCount,
			d_x.Milliseconds(), d_dchest.Milliseconds())
	}
}

// blake2b (dchest to x) migration tests
func TestBlake2b_256(t *testing.T) {
	sampleHash32 := [32]byte{}
	hash_curr := make([]byte, 32)
	HashCompress(hash_curr, sampleHash32[:]) //current implementation
	hash_old := make([]byte, 32)
	HashCompress_Old(hash_old, sampleHash32[:]) //new implementation

	if !bytes.Equal(hash_curr, hash_old) {
		t.Errorf("%s - mismatch %02X <-> %02X", t.Name(), hash_curr, hash_old)
	}
}

func TestBlake2b_512(t *testing.T) {
	sampleHash64 := [64]byte{}
	hash_curr := make([]byte, 64)
	HashCompress64(hash_curr, sampleHash64[:]) //current implementation
	hash_new := make([]byte, 64)
	HashCompress64_Old(hash_new, sampleHash64[:]) //new implementation
	if !bytes.Equal(hash_curr, hash_new) {
		t.Errorf("%s - mismatch %02X <-> %02X", t.Name(), hash_curr, hash_new)
	}
}

// these are the old dchest methods
func HashCompress_Old(out, in []byte) {
	if len(out) < 32 {
		panic("need 32 byte output to place hash in")
	}
	b2 := b2b_dchest.New256()
	_, err := b2.Write(in)
	if err != nil {
		panic("failed b2.Write()")
	}
	// blake2 wants to *append* the hash
	b2.Sum(out[:0])
}

func HashCompress64_Old(out, in []byte) {
	if len(out) < 64 {
		panic("need 64 byte output to place hash in")
	}
	b2 := b2b_dchest.New512()
	_, err := b2.Write(in)
	if err != nil {
		panic("failed b2.Write()")
	}
	// blake2 wants to *append* the hash
	b2.Sum(out[:0])
}
