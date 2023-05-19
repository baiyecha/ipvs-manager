package utils

import (
	"baiyecha/ipvs-manager/model"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/hashicorp/raft"
)

func GetLeader(address []string) string {
	// fmt.Println("address", address)
	leaderAddr := ""
	for _, addr := range address {

		// 发起GET请求
		resp, err := http.Get(fmt.Sprintf("http://%s/raft/stats", addr))
		if err != nil {
			// 处理错误
			fmt.Println("请求错误:", err)
			continue
		}

		defer resp.Body.Close()

		// 读取响应的body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// 处理错误
			fmt.Println("读取body错误:", err)
			continue
		}
		rsResp := &model.RaftStatsResp{}
		err = json.Unmarshal(body, rsResp)
		if err != nil {
			fmt.Println(err)
			continue
		}
		// fmt.Printf("GetLeader res: status: %d , data: %+v \n",res.StatusCode, resp)
		if rsResp.Data.State == raft.Leader.String() {
			// fmt.Println("leader is ", addr)
			leaderAddr = addr
			break
		}
	}
	return leaderAddr
}

func PostRequest(url string, data interface{}) (string, int, error) {
	// 将data参数转换为JSON格式字符串
	payload, err := json.Marshal(data)
	if err != nil {
		return "", 0, err
	}

	// 创建POST请求
	req, err := http.NewRequest("POST", url, strings.NewReader(string(payload)))
	if err != nil {
		return "", 0, err
	}

	// 设置请求头部信息
	req.Header.Set("Content-Type", "application/json")

	// 创建自定义的Transport，支持非安全的HTTPS请求
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	defer tr.CloseIdleConnections()

	// 发起请求
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}

	defer resp.Body.Close()

	// 读取响应的body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", 0, err
	}

	// 返回响应内容和状态码
	return string(body), resp.StatusCode, nil
}

func GetRequest(url string) (string, int, error) {
	// 创建自定义的Transport，支持非安全的HTTPS请求
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	defer tr.CloseIdleConnections()

	// 创建带有自定义Transport的http.Client
	client := &http.Client{Transport: tr}

	// 发起GET请求
	resp, err := client.Get(url)
	if err != nil {
		return "", 0, err
	}

	defer resp.Body.Close()

	// 读取响应的body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", 0, err
	}

	// 返回响应内容和状态码
	return string(body), resp.StatusCode, nil
}
