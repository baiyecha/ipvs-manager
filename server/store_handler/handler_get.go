package store_handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"baiyecha/ipvs-manager/constant"
	pb "baiyecha/ipvs-manager/grpc/proto"
	"baiyecha/ipvs-manager/model"
	"baiyecha/ipvs-manager/utils"

	"github.com/dgraph-io/badger/v2"
	"github.com/labstack/echo/v4"
	"github.com/levigross/grequests"
)

// Get will fetched data from badgerDB where the raft use to store data.
// It can be done in any raft server, making the Get returned eventual consistency on read.
func (h handler) Get(eCtx echo.Context) error {
	key := strings.TrimSpace(eCtx.Param("key"))
	if key == "" {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": "key is empty",
		})
	}

	keyByte := []byte(key)

	txn := h.db.NewTransaction(false)
	defer func() {
		_ = txn.Commit()
	}()

	item, err := txn.Get(keyByte)
	if err != nil {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": fmt.Sprintf("error getting key %s from storage: %s", key, err.Error()),
		})
	}

	value := make([]byte, 0)
	err = item.Value(func(val []byte) error {
		value = append(value, val...)
		return nil
	})

	if err != nil {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": fmt.Sprintf("error appending byte value of key %s from storage: %s", key, err.Error()),
		})
	}

	var data interface{}
	if value != nil && len(value) > 0 {
		err = json.Unmarshal(value, &data)
	}

	if err != nil {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": fmt.Sprintf("error unmarshal data to interface: %s", err.Error()),
		})
	}

	return eCtx.JSON(http.StatusOK, map[string]interface{}{
		"message": "success fetching data",
		"data": map[string]interface{}{
			"key":   key,
			"value": data,
		},
	})
}

func (h handler) ServiceInfo(eCtx echo.Context) error {
	// ServiceInfo页面
	// todo
	si := model.ServiceInfo{}
	getStore(h.db, constant.NodeStatusKey, &si)
	for _, s := range si.Servers {
		if isWithinHour(s.LastHeartTime) {
			s.Status = "down"
		} else {
			s.Status = "up"
		}
	}
	for _, a := range si.Agents {
		if isWithinHour(a.LastHeartTime) {
			a.Status = "down"
		} else {
			a.Status = "up"
		}
	}
	return eCtx.Render(http.StatusOK, "agent.html", &si)
}

func isWithinHour(timestamp string) bool {
	// 解析时间字符串
	t, err := time.Parse("2006-01-02 15:04:05", timestamp)
	if err != nil {
		// 解析失败，返回 false
		return true
	}

	// 计算与当前时间的差值
	diff := time.Now().Sub(t)

	// 判断是否在一个小时之前
	if diff > time.Hour {
		return true
	} else {
		return false
	}
}

func (h handler) Table(eCtx echo.Context) error {
	// table页面
	//cookie, err := eCtx.Cookie(constant.CookieName)
	//if err != nil || cookie.Value != constant.NameAndPwd {
	//	return eCtx.Redirect(http.StatusMovedPermanently, "/")
	//}
	// 先拿出所有的ipvs信息
	ipvsList := &model.IpvsList{}
	getStore(h.db, constant.IpvsStroreKey, ipvsList)
	if len(ipvsList.List) == 0 {
		ipvsList.List = make([]*model.Ipvs, 0)
	}
	jsonStr, _ := json.Marshal(ipvsList.List)
	ipvsList.Json = string(jsonStr)
	return eCtx.Render(http.StatusOK, "table.html", ipvsList)
}

func getStore(db *badger.DB, k string, v interface{}) error {
	txn := db.NewTransaction(false)
	defer func() {
		_ = txn.Commit()
	}()
	item, err := txn.Get([]byte(k))
	value := make([]byte, 0)
	if err != nil || item == nil {
		fmt.Print(err)
	} else {
		err = item.Value(func(val []byte) error {
			value = append(value, val...)
			return nil
		})
		if len(value) > 0 {
			err = json.Unmarshal(value, v)
		}
		if err != nil {
			fmt.Printf("error unmarshal data to interface: %s", err.Error())
			return err
		}
	}
	return nil
}

func (gss *GrpcStoreServer) IpvsList(ctx context.Context, request *pb.IpvsListRequeste) (*pb.IpvsListResponse, error) {
	// 这里先记录一下心跳
	Heartbeat(gss.db, &model.NodeInfo{
		IP: request.Ip,
	}, constant.AgentRule, gss.clusterAddress)
	// 先拿出所有的ipvs信息
	txn := gss.db.NewTransaction(false)
	defer func() {
		_ = txn.Commit()
	}()
	ipvsList := &pb.IpvsListResponse{}
	item, err := txn.Get([]byte(constant.IpvsStroreKey))
	value := make([]byte, 0)
	if err != nil || item == nil {
		if err == badger.ErrKeyNotFound {
			return ipvsList, nil
		} else {
			fmt.Print(err)
		}
	} else {
		err = item.Value(func(val []byte) error {
			value = append(value, val...)
			return nil
		})
		if len(value) > 0 {
			err = json.Unmarshal(value, ipvsList)
		}
		if err != nil {
			fmt.Printf("error unmarshal data to interface: %s", err.Error())
		}
	}
	return ipvsList, err
}

func Heartbeat(db *badger.DB, nodeInfo *model.NodeInfo, rule string, clusterAddress []string) error {
	if nodeInfo.IP == "" {
		return nil
	}
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		fmt.Println("获取时区失败:", err)
		return err
	}
	t := time.Now().In(location)
	currentTimeString := t.Format("2006-01-02 15:04:05")
	serverInfo := &model.ServiceInfo{}
	err = getStore(db, constant.NodeStatusKey, serverInfo)
	if err != nil {
		if err != badger.ErrKeyNotFound {
			return err
		}
	}

	switch rule {
	case constant.AgentRule:
		isadd := true
		for _, agent := range serverInfo.Agents {
			if agent.IP == nodeInfo.IP {
				agent.LastHeartTime = currentTimeString
				isadd = false
			}
		}
		if isadd {
			serverInfo.Agents = append(serverInfo.Agents, nodeInfo)
		}
	case constant.ServerRule:
		isadd := true
		for _, server := range serverInfo.Servers {
			if server.IP == nodeInfo.IP {
				server.LastHeartTime = currentTimeString
				server.IsLeader = nodeInfo.IsLeader
				isadd = false
			}
		}
		if isadd {
			nodeInfo.LastHeartTime = currentTimeString
			serverInfo.Servers = append(serverInfo.Servers, nodeInfo)
		}
	}
	// 更新 node状态
	// 如果不是leader ，则发送请求给leader进行操作
	leaderAddr := utils.GetLeader(clusterAddress)
	if leaderAddr == "" {
		return fmt.Errorf("not found leader")
	}
	res, err := grequests.Post(fmt.Sprintf("http://%s/store/", leaderAddr), &grequests.RequestOptions{
		JSON: requestStore{
			Key:   constant.NodeStatusKey,
			Value: serverInfo,
		},
	})
	if err != nil {
		fmt.Print(err)
		return err
	}
	if res.StatusCode != 200 {
		fmt.Print(res.StatusCode)
	}
	return nil
}
