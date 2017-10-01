package assembler

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/thee-engineer/cryptor/cachedb"
	"github.com/thee-engineer/cryptor/chunker"
	"github.com/thee-engineer/cryptor/crypt"
)

func TestEChunk(t *testing.T) {
	// Create temporary dir for test
	tmpDir, err := ioutil.TempDir("/tmp", "assembler")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create temp cache
	cache, err := cachedb.NewLDBCache(tmpDir, 0, 0)
	if err != nil {
		t.Error(err)
	}

	// Test data
	var buffer bytes.Buffer
	data := crypt.RandomData(520)
	buffer.Write(data)

	// Create chunker
	chunker := &chunker.Chunker{
		Size:   1024,
		Cache:  cache,
		Reader: &buffer,
	}
	chunkHash, err := chunker.Chunk(crypt.NullKey)
	if err != nil {
		t.Error(err)
	}

	// Read encrypted chunk
	eChunk := GetEChunk(chunkHash, cache)
	dChunk, err := eChunk.Decrypt(crypt.NullKey)
	if err != nil {
		t.Error(err)
	}

	// Invalid hash
	if !dChunk.IsValid() {
		t.Error("Chunk is not valid!")
	}

	// Chunk should be the tail (as it is the only chunk)
	if !dChunk.IsLast() {
		t.Error("Chunk is not tail!")
	}

	// Compare initial data with data after encryption, storage and decryption
	if !bytes.Equal(dChunk.Content, data) {
		t.Error("Data mismatch!")
	}
}