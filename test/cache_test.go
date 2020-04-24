package test

import (
	"fmt"
	"myGoService/model"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	cd := model.GetCacheDriver()
	if cd == nil {
		fmt.Println("a")
	}
	value := "yahoo"
	cd.SetCacheWithKeyVal("test", value)
	time.Sleep(1 * time.Second)
	fmt.Println(cd.GetCacheValWithKey("test"))
}
