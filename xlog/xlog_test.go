package xlog

import (
	"os"
	"testing"
)

func TestSetLevel(t *testing.T) {
	SetLevel(InfoLevel, os.Stdout)
	SetLevel(ErrorLevel, os.Stdout)

}

func TestColor(t *testing.T){
	Infoln("The system starts...")
	Printf("The systems returns with code %x\n", 0x01)
	Errorf("the RPC server port crashes (:%d)\n", 9999)
	Fatalln("Oops, the system shut down for unknown reason")
}