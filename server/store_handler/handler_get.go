package store_handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"baiyecha/ipvs-manager/constant"
	pb "baiyecha/ipvs-manager/grpc/proto"
	"baiyecha/ipvs-manager/model"

	"github.com/dgraph-io/badger/v2"
	"github.com/labstack/echo/v4"
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
	si := model.ServiceInfo{
		Servers: []*model.NodeInfo{
			{
				IP:            "1.1.1.1",
				RpcPort:       "80",
				WebPort:       "8000",
				IsLeader:      "是",
				LastHeartTime: "2023/1/1",
				Status:        "正常",
			},
			{
				IP:            "1.1.1.2",
				RpcPort:       "81",
				WebPort:       "8001",
				IsLeader:      "是",
				LastHeartTime: "2023/1/2",
				Status:        "异常",
			},
		},
		Agents: []*model.NodeInfo{
			{
				IP:            "1.1.1.3",
				RpcPort:       "82",
				WebPort:       "8002",
				IsLeader:      "是",
				LastHeartTime: "2023/1/3",
				Status:        "异常",
			},
		},
	}
	return eCtx.Render(http.StatusOK, "agent.html", &si)
}

func (h handler) Table(eCtx echo.Context) error {
	// table页面
	//cookie, err := eCtx.Cookie(constant.CookieName)
	//if err != nil || cookie.Value != constant.NameAndPwd {
	//	return eCtx.Redirect(http.StatusMovedPermanently, "/")
	//}
	// 先拿出所有的ipvs信息
	txn := h.db.NewTransaction(false)
	ipvsList := &model.IpvsList{}
	item, err := txn.Get([]byte(constant.IpvsStroreKey))
	value := make([]byte, 0)
	if err != nil || item == nil {
		fmt.Print(err)
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
	if len(ipvsList.List) == 0 {
		ipvsList.List = make([]*model.Ipvs, 0)
	}
	jsonStr, _ := json.Marshal(ipvsList.List)
	ipvsList.Json = string(jsonStr)
	return eCtx.Render(http.StatusOK, "table.html", ipvsList)
}

func (gss *GrpcStoreServer) IpvsList(ctx context.Context, request *pb.IpvsListRequeste) (*pb.IpvsListResponse, error) {
	// 先拿出所有的ipvs信息
	txn := gss.db.NewTransaction(false)
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
