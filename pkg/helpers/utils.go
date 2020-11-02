package helpers

import (
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyz"

// WaitOnErrorChannel waits till channel closed
func WaitOnErrorChannel(ch chan error, wg *sync.WaitGroup) error {

	go func() {
		defer close(ch)
		wg.Wait()
	}()

	result := &multierror.Error{}

	for err := range ch {
		result = multierror.Append(result, err)
	}

	return result.ErrorOrNil()

}

// GetPrefix returns prefix for resources
func GetPrefix(prefix string) string {
	return prefix + "-"
}

// LastPartOnSplit splits string on delimiter and returns last part
func LastPartOnSplit(s, delimiter string) string {
	return s[strings.LastIndex(s, delimiter)+1:]
}

// FilterStrings filters in place strings slice
func FilterStrings(items *[]string, handler func(item string) bool) {

	start := 0
	for i := start; i < len(*items); i++ {
		if !handler((*items)[i]) {
			// vm will be deleted
			continue
		}
		if i != start {
			(*items)[start], (*items)[i] = (*items)[i], (*items)[start]
		}
		start++
	}

	*items = (*items)[:start]

}

// RandStringBytes generates random string with defined length
func RandStringBytes(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		// nolint
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func FindStrIndex(input string, search []string) int {
	for idx, loc := range search {
		if loc == input {
			return idx
		}
	}
	return -1
}

func StringsContains(input string, search []string) (int, bool) {
	idx := FindStrIndex(input, search)
	return idx, idx != -1
}

func RemoveFromSlice(slice []string, i int) []string {
	slice[i] = slice[len(slice)-1]
	slice[len(slice)-1] = ""
	slice = slice[:len(slice)-1]
	return slice
}

func SortIntSPosition(names []string, positions []string, values []int) []int {
	res := make([]int, len(values))
	for idx, name := range names {
		posIdx := FindStrIndex(name, positions)
		res[posIdx] = values[idx]
	}
	return res
}
