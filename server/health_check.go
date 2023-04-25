package server

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"runtime/debug"
	"time"

	"baiyecha/ipvs-manager/constant"
	"baiyecha/ipvs-manager/model"
	"baiyecha/ipvs-manager/server/store_handler"

	"github.com/dgraph-io/badger/v2"
	"github.com/hashicorp/raft"
	"github.com/levigross/grequests"
)

func RunHealthCheck(badgerDB *badger.DB, r *raft.Raft) {
	fmt.Print("run health check...")
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
					continue
				}
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
					fmt.Printf("error appending byte value of key ipvs from storage: %s", err.Error())
					goto done
				}
				if len(value) > 0 {
					err = json.Unmarshal(value, ipvsList)
				}
				if err != nil {
					fmt.Printf("error unmarshal data to interface: %s", err.Error())
					goto done
				}
				// 检测所有ipvs 后端是否存活
				_, isChange = doHealthCheck(ipvsList)
				if isChange {
					err := store_handler.Store(r, "ipvs", ipvsList)
					if err != nil {
						fmt.Printf("update ipvs data error: %s", err.Error())
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
				status = httpCheck(backend.CheckInfo)
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
	conn, err := net.Dial(protocol, host+":"+port)
	if err != nil {
		fmt.Printf("Port %s is closed\n", host+":"+port)
		fmt.Println(err)
		return 1
	} else {
		fmt.Printf("Port %s is open\n", host+":"+port)
		conn.Close()
		return 0
	}
}

func httpCheck(url string) int {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	res, err := grequests.Get(url, &grequests.RequestOptions{
		HTTPClient: &http.Client{
			Transport: tr,
		},
	})
	if err != nil {
		fmt.Print(err)
		return 1
	}
	if res.StatusCode == http.StatusOK {
		fmt.Printf("check %s is ok!\n", url)
		return 0
	} else {
		fmt.Printf("check %s is failed! status: %d\n", url, res.StatusCode)
		return 1
	}
}
