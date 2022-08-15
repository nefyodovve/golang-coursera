package main

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

func ExecutePipeline(jobs ...job) {
	wg := &sync.WaitGroup{}
	in := make(chan interface{})
	for _, j := range jobs {
		out := make(chan interface{})
		wg.Add(1)
		go func(in, out chan interface{}, j job) {
			j(in, out)
			close(out)
			wg.Done()
		}(in, out, j)

		in = out
	}
	wg.Wait()

}

func SingleHash(in, out chan interface{}) {
	type md5Result struct {
		data string
		md5  string
	}
	var md5Results []md5Result
	for data := range in {
		dataStr := fmt.Sprint(data.(int))
		md5Results = append(md5Results, md5Result{data: dataStr, md5: DataSignerMd5(dataStr)})
	}

	var hashSlice []chan string
	for _, r := range md5Results {
		hashSlice = append(hashSlice, execInGoroutine(DataSignerCrc32, r.data))
		hashSlice = append(hashSlice, execInGoroutine(DataSignerCrc32, r.md5))
	}
	result := ""
	second := false
	for _, hash := range hashSlice {
		if second {
			result += "~" + <-hash
			out <- result
			second = false
			result = ""
		} else {
			result += <-hash
			second = true
		}
	}
}

func MultiHash(in, out chan interface{}) {
	th := 6
	var hashSlice []chan string
	for data := range in {
		for i := 0; i < th; i++ {
			hashSlice = append(hashSlice, execInGoroutine(DataSignerCrc32, fmt.Sprint(i)+data.(string)))
		}
	}
	result := ""
	for i, hash := range hashSlice {
		if i != 0 && i%th == 0 {
			out <- result
			result = <-hash
		} else {
			result += <-hash
		}
	}
	out <- result
}

func CombineResults(in, out chan interface{}) {
	var allData []string
	for data := range in {
		allData = append(allData, data.(string))
	}
	sort.Strings(allData)
	out <- strings.Join(allData, "_")
}

func execInGoroutine(f func(data string) string, data string) chan string {
	result := make(chan string, 1)
	go func(out chan<- string) {
		out <- f(data)
		close(out)
	}(result)
	return result
}
