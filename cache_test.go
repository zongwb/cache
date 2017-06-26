package cache

import (
	"fmt"
	"hash/crc32"
	"os"
	"testing"
)

// HashStringCRC32 hashes a string using CRC32 algorithm.
func HashStringCRC32(key interface{}) uint32 {
	k := key.(string)
	return crc32.ChecksumIEEE([]byte(k))
}

func TestCache(t *testing.T) {
	fmt.Println("\nTesting single LRUCache")
	sz := 2
	cache := NewLRUCache(sz)
	c := cache.(*LRUCache)
	tab := []struct {
		key string
		val int
	}{
		{"A", 1},
		{"B", 2},
		{"C", 3},
		{"D", 4},
	}
	for _, e := range tab {
		c.Set(e.key, e.val)
	}
	c.PrintAll(os.Stdout, "\n")
}

func TestCombo(t *testing.T) {
	fmt.Println("\nTesting ComboLRUCache")
	combo := NewComboLRUCache(10, 2, HashStringCRC32)
	c := combo.(*ComboLRUCache)
	c.Set("A", 1)
	tab := []struct {
		key string
		val int
	}{
		{"A", 1},
		{"B", 2},
		{"C", 3},
		{"D", 4},
		{"E", 5},
		{"F", 6},
		{"G", 7},
		{"H", 8},
		{"I", 9},
		{"J", 10},
		{"K", 11},
		{"L", 12},
		{"M", 13},
		{"N", 14},
	}
	for _, e := range tab {
		c.Set(e.key, e.val)
	}
	c.PrintAll(os.Stdout, "\n")
}
