package main

import (
	"fmt"
	"math/rand"
)

func deleteElement[T any](index int, s []T) []T {
	// 如果下标不正确，则抛出错误
	if index < 0 || index >= len(s) {
		panic("Index error")
	}
	newSlice := append(s[:index], s[index+1:]...)
	// 如果除于 2 后的容量还是太大，则缩容
	if cap(newSlice)/2 >= len(newSlice) {
		shrinkSlice := make([]T, cap(newSlice)/2)
		copy(shrinkSlice, newSlice)
		return shrinkSlice
	}
	return newSlice
}

func main() {
	s := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	fmt.Printf("%v len: %d cap: %d\n", s, len(s), cap(s))

	// 进行 7 次删除
	for i := 0; i < 7; i++ {
		// 删除随机下标
		randIndex := rand.Intn(len(s))
		s = deleteElement[int](randIndex, s)
		fmt.Printf("Delete index %d %v len: %d cap: %d\n", randIndex, s, len(s), cap(s))
	}
}
