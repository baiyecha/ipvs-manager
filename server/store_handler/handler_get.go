package store_handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"baiyecha/ipvs-manager/constant"
	"baiyecha/ipvs-manager/model"
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

func (h handler) Table(eCtx echo.Context) error {
	// table页面
	cookie, err := eCtx.Cookie(constant.CookieName)
	if err != nil || cookie.Value != constant.NameAndPwd {
		return eCtx.Redirect(http.StatusMovedPermanently, "/")
	}
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
	if len(ipvsList.IpvsList) == 0 {
		ipvsList.IpvsList = make([]*model.Ipvs, 0)
	}
	jsonStr, _ := json.Marshal(ipvsList.IpvsList)
	ipvsList.Json = string(jsonStr)
	return eCtx.Render(http.StatusOK, "table.html", ipvsList)
}
