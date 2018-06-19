package cache

import (
	"fmt"
	"testing"
	"time"
)

func Test_cache(t *testing.T) {
	lruCache := NewCache(100, 1*time.Second)
	lruCache.Set("key", "Hello, world!", 3*time.Second)
	res, ok := lruCache.Get("key")
	if ok {
		fmt.Printf("result is %+v\n", res)
	} else {
		fmt.Println("key does not exist")
	}
	time.Sleep(5 * time.Second)
	res, ok = lruCache.Get("key")
	if ok {
		fmt.Printf("result is %+v\n", res)
	} else {
		fmt.Println("key does not exist")
	}
}
