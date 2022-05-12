package grasure

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

func TestGetLocalAddr(t *testing.T) {
	ret := getLocalAddr()
	fmt.Println(ret)
}

func TestGenUUID(t *testing.T) {
	for i := 0; i < 100; i++ {
		fmt.Println(genUUID(1))
	}
}

//show local disks with `ls /dev/sd*`
//and use `df` to check
func TestDiskUsage(t *testing.T) {
	path := "/dev/sda"
	total, free := DiskUsage(path)
	out, err := exec.Command("df", path).Output()
	if err != nil {
		t.Fatal(err)
	}
	reg := regexp.MustCompile(`[0-9]+(\s)`)
	res := reg.FindAllString(string(out), -1)
	for i := range res {
		res[i] = strings.Split(res[i], " ")[0]
	}
	//the `df` shows in 1024 blocks
	if res[0] != toString(total/1024) || res[2] != toString(free/1024) {
		t.Fatal("fail")
	}
}
