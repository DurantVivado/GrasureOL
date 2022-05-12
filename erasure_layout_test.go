package grasure

import (
	"fmt"
	"testing"
)

func TestLayout_Random(t *testing.T) {
	layout := NewLayout(Random, 100, 4, 1, 10)
	fmt.Println(layout.distribution)
	fmt.Println(layout.blockToOffset)
}

func TestLayout_LeftAsymmetric(t *testing.T) {
	layout := NewLayout(LeftAsymmetric, 100, 4, 2, 10)
	fmt.Println(layout.distribution)
	fmt.Println(layout.blockToOffset)
}

func TestLayout_LeftSymmetric(t *testing.T) {
	layout := NewLayout(LeftSymmetric, 100, 4, 2, 10)
	fmt.Println(layout.distribution)
	fmt.Println(layout.blockToOffset)
}

func TestLayout_RightAsymmetric(t *testing.T) {
	layout := NewLayout(RightAsymmetric, 100, 4, 2, 10)
	fmt.Println(layout.distribution)
	fmt.Println(layout.blockToOffset)
}

func TestLayout_RightSymmetric(t *testing.T) {
	layout := NewLayout(RightSymmetric, 100, 4, 2, 10)
	fmt.Println(layout.distribution)
	fmt.Println(layout.blockToOffset)
}
