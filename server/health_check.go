package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
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
	for ipvsDataIndex := range ipvsList.IpvsList {
		ipvsData := ipvsList.IpvsList[ipvsDataIndex]
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
	timeout := time.Second
	conn, err := net.DialTimeout(protocol, host+":"+port, timeout)
	if err != nil {
		fmt.Println(err)
		return 1
	} else {
		defer conn.Close()
		msg, _, err := bufio.NewReader(conn).ReadLine()
		if err != nil {
			if err == io.EOF {
				fmt.Print(host + "" + port + " - Open!")
				return 0
			}
		} else {
			fmt.Print(host + "" + port + " -" + string(msg))
			return 1
		}
	}
	return 1
}

func httpCheck(url string) int {
	res, err := grequests.Get(url, &grequests.RequestOptions{})
	if err != nil {
		fmt.Print(err)
		return 1
	}
	if res.StatusCode == http.StatusOK {
		fmt.Printf("check %s is ok!", url)
		return 0
	} else {
		fmt.Printf("check %s is failed! status: %d", url, res.StatusCode)
		return 1
	}
	return 1
}
