package server

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"runtime/debug"
	"strconv"
	"time"

	"baiyecha/ipvs-manager/constant"
	"baiyecha/ipvs-manager/model"
	"baiyecha/ipvs-manager/server/store_handler"
	"baiyecha/ipvs-manager/utils"

	"github.com/dgraph-io/badger/v2"
	"github.com/hashicorp/raft"
)

func RunHealthCheck(badgerDB *badger.DB, r *raft.Raft) {
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Println("consumer task error", "err", r, "stack", string(debug.Stack()))
				}
			}()
			for {
				// 定时请求进行心跳检测
				// 先判断自身节点是否为leader
				if r.State() != raft.Leader {
					time.Sleep(5 * time.Second)
					continue
				}
				fmt.Println("run health check...")
				// 检测心跳
				// 先拿出所有的ipvs信息
				txn := badgerDB.NewTransaction(false)
				ipvsList := &model.IpvsList{}
				item, err := txn.Get([]byte(constant.IpvsStroreKey))
				value := make([]byte, 0)
				isChange := false
				if err != nil {
					fmt.Print(err)
					goto done
				}
				err = item.Value(func(val []byte) error {
					value = append(value, val...)
					return nil
				})

				if err != nil {
					fmt.Printf("error appending byte value of key ipvs from storage: %s \n", err.Error())
					goto done
				}
				if len(value) > 0 {
					err = json.Unmarshal(value, ipvsList)
				}
				if err != nil {
					fmt.Printf("error unmarshal data to interface: %s \n", err.Error())
					goto done
				}
				// 检测所有ipvs 后端是否存活
				_, isChange = doHealthCheck(ipvsList)
				if isChange {
					err := store_handler.Store(r, constant.IpvsStroreKey, ipvsList)
					if err != nil {
						fmt.Printf("update ipvs data error: %s \n", err.Error())
						goto done
					}
				}
			done:
				_ = txn.Commit()
				time.Sleep(5 * time.Second)
			}
		}()
	}
}

func doHealthCheck(ipvsList *model.IpvsList) (error, bool) {
	isChange := false
	for ipvsDataIndex := range ipvsList.List {
		ipvsData := ipvsList.List[ipvsDataIndex]
		for backendIndex := range ipvsData.Backends {
			backend := ipvsData.Backends[backendIndex]
			status := 1
			switch backend.CheckType {
			case 0: // tcp
				addr := backend.Addr
				if backend.CheckInfo != "" {
					addr = backend.CheckInfo
				}
				ip, port, _ := net.SplitHostPort(addr)
				status = telnet(ipvsData.Protocol, ip, port)
			case 1: // http
				status = httpCheck(backend.CheckInfo, backend.CheckResType, backend.CheckRes)
			default:
				backend.Status = 1
			}
			if status != backend.Status {
				backend.Status = status
				isChange = true
			}
		}
	}
	return nil, isChange
}

// @protocol tcp or udp
// @return 0 succeed 1 failed
func telnet(protocol string, host string, port string) int {
	conn, err := net.DialTimeout(protocol, host+":"+port, 500*time.Millisecond)
	if err != nil {
		fmt.Printf("Port %s is closed\n", host+":"+port)
		fmt.Println(err)
		return 1
	} else {
		conn.Close()
		// fmt.Printf("Port %s is open\n", host+":"+port)
		return 0
	}
}

func httpCheck(url string, checkResType int, checkRes string) int {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	defer tr.CloseIdleConnections()

	res, statusCode, err := utils.GetRequest(url)

	if err != nil {
		fmt.Println(err)
		return 1
	}
	switch checkResType {
	case 0:
		if checkRes == "" {
			checkRes = "200"
		}
		return isMatch(strconv.Itoa(statusCode), checkRes)
	case 1:
		if checkRes == "" {
			checkRes = "ok"
		}
		return isMatch(res, checkRes)
	}
	return 1
}

func isMatch(str string, pattern string) int {

	// 编译正则表达式
	reg := regexp.MustCompile(pattern)

	// 使用正则表达式匹配字符串
	result := reg.FindAllString(str, -1)

	if len(result) == 0 {
		return 1
	} else {
		return 0
	}
}
