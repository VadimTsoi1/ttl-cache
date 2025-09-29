package main

import (
	"fmt"
	"time"

	"github.com/VadimTsoi1/ttl-cache/cache"
)

func main() {
	c := cache.NewCache(100 * time.Millisecond)
	defer c.Close()
	c.CleanUp()

	c.Set("name", "Vadim", 200*time.Millisecond)

	if v, ok := c.Get("name"); ok {
		fmt.Println("immediate:", v)
	}

	time.Sleep(300 * time.Millisecond)

	if _, ok := c.Get("name"); !ok {
		fmt.Println("expired and removed(by Get or CleanUp)")
	}
}
