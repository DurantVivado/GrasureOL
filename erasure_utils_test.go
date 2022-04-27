package grasure

import (
	"fmt"
	"testing"
)

func TestGetLocalAddr(t *testing.T){
	ret := getLocalAddr()
	fmt.Println(ret)
}

func TestGenUUID(t *testing.T){
	for i:=0;i<100;i++ {
		fmt.Println(genUUID(1))
	}
}