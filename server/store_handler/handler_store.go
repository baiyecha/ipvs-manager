package store_handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/raft"
	"github.com/labstack/echo/v4"
	"github.com/levigross/grequests"

	"baiyecha/ipvs-manager/constant"
	"baiyecha/ipvs-manager/fsm"
	"baiyecha/ipvs-manager/model"
	"baiyecha/ipvs-manager/utils"
)

// requestStore payload for storing new data in raft cluster
type requestStore struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

// Store handling save to raft cluster. Store will invoke raft.Apply to make this stored in all cluster
// with acknowledge from n quorum. Store must be done in raft leader, otherwise return error.
func (h handler) Store(eCtx echo.Context) error {
	form := requestStore{}
	if err := eCtx.Bind(&form); err != nil {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": fmt.Sprintf("error binding: %s", err.Error()),
		})
	}

	form.Key = strings.TrimSpace(form.Key)
	if form.Key == "" {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": "key is empty",
		})
	}

	if h.raft.State() != raft.Leader {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": "not the leader",
		})
	}

	err := Store(h.raft, form.Key, form.Value)
	if err != nil {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": err,
		})
	}

	return eCtx.JSON(http.StatusOK, map[string]interface{}{
		"message": "success persisting data",
		"data":    form,
	})
}

func (h handler) Update(eCtx echo.Context) error {
	form := model.IpvsList{}
	if err := eCtx.Bind(&form); err != nil {
		fmt.Print(err)
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": fmt.Sprintf("error binding: %s", err.Error()),
		})
	}
	if h.raft.State() == raft.Leader {
		err := Store(h.raft, constant.IpvsStroreKey, form)
		if err != nil {
			fmt.Print("111", err)
			return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
				"error": err,
			})
		}
		return eCtx.JSON(http.StatusOK, map[string]interface{}{
			"message": "success persisting data",
			"data":    form,
		})
	}
	// 如果不是leader ，则发送请求给leader进行操作
	leaderAddr := utils.GetLeader(h.clusterAddress)
	if leaderAddr == "" {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": fmt.Sprintf("not found leader"),
		})
	}
	res, err := grequests.Post(fmt.Sprintf("http://%s/store/",leaderAddr), &grequests.RequestOptions{
		JSON: requestStore{
			Key:   constant.IpvsStroreKey,
			Value: form,
		},
	})
	if err != nil {
		fmt.Print(err)
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": err,
		})
	}
	if res.StatusCode != 200 {
		fmt.Print(res.StatusCode)
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": err,
		})
	}
	return eCtx.JSON(http.StatusOK, map[string]interface{}{
		"message": "success persisting data",
		"data":    form,
	})
}

func Store(r *raft.Raft, key string, value interface{}) error {
	payload := fsm.CommandPayload{
		Operation: "SET",
		Key:       key,
		Value:     value,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error preparing saving data payload: %s", err.Error())
	}

	applyFuture := r.Apply(data, 500*time.Millisecond)
	if err := applyFuture.Error(); err != nil {
		return fmt.Errorf("error persisting data in raft cluster: %s", err.Error())
	}

	_, ok := applyFuture.Response().(*fsm.ApplyResponse)
	if !ok {
		return fmt.Errorf("error response is not match apply response")
	}

	return nil
}
