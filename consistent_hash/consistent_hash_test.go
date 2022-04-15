package consistentHash

import (
	"strconv"
	"testing"
)

var testCasesAscii = map[string]string{
	"1":  "2",
	"2":  "2",
	"3":  "4",
	"4":  "4",
	"5":  "6",
	"15": "6",
	"20": "2",
	"30": "2",
	"40": "2",
	"50": "2",
}

var testCasesMD5 = map[string]string{
	"1":  "2",
	"2":  "2",
	"3":  "4",
	"4":  "4",
	"5":  "6",
	"15": "6",
	"20": "2",
	"30": "2",
	"40": "2",
	"50": "2",
}

var testCasesCRC = map[string]string{
	"1":  "2",
	"2":  "2",
	"3":  "4",
	"4":  "4",
	"5":  "6",
	"15": "6",
	"20": "2",
	"30": "2",
	"40": "2",
	"50": "2",
}

func TestConsistentHashAscii(t *testing.T) {
	hash := NewConsistentHash(3, func(key []byte) uint32 {
		ret, _ := strconv.Atoi(string(key))
		return uint32(ret)
	})
	//Given the above hash func, this will give replica wiht hashes:
	//2, 4, 6, 12, 14, 16, 22, 24, 26
	hash.AddNode("2", "4", "6")
	for k, v := range testCasesAscii {
		if ret := hash.GetNode(k); ret != v {
			t.Errorf("k:%s,Should be %s, but yielded %s", k, v, ret)
		}
	}
	hash.AddNode("8")
	testCasesAscii["27"] = "8"
	for k, v := range testCasesAscii {
		if ret := hash.GetNode(k); ret != v {
			t.Errorf("k:%s,Should be %s, but yielded %s", k, v, ret)
		}
	}

	hash.DelNode("8")
	testCasesAscii["27"] = "2"
	for k, v := range testCasesAscii {
		if ret := hash.GetNode(k); ret != v {
			t.Errorf("k:%s,Should be %s, but yielded %s", k, v, ret)
		}
	}

}
