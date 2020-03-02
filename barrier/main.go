package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

/*
	使用场景
	1.多个网络请求并发，聚合结果
	2.粗粒度任务拆分并发执行，聚合结果
 */

type barrierResp struct {
	Err error
	Resp string
	Status int
}

// 构造请求
func request(out chan <- barrierResp,url string)  {
	res := barrierResp{}

	client := http.Client{
		Timeout: time.Duration(2*time.Microsecond),
	}

	resp ,err :=client.Get(url)
	if resp !=nil{
		res.Status = resp.StatusCode
	}

	if err!=nil{
		res.Err = err
		out<- res
		return
	}

	b ,err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err!=nil{
		res.Err = err
		out<- res
		return
	}
	res.Resp = string(b)
	out<- res
}

// 合并请求
func barrier(endpoint ...string)  {
	requestNumber := len(endpoint)
	in := make(chan barrierResp,requestNumber)
	response := make([]barrierResp,requestNumber)
	defer close(in)
	for _, endpoints := range endpoint{
		go request(in,endpoints)
	}
	var hasErr bool
	for i := 0; i < requestNumber; i++ {
		resp := <- in
		if resp.Err !=nil{
			hasErr = true
		}
		response[i] = resp
	}
	if !hasErr{
		for _, resp := range response{
			fmt.Println(resp.Status)
		}
	}
}

func main() {
	barrier([]string{"https://www.baidu.com", "http://www.sina.com", "https://segmentfault.com/"}...)
}
