package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kagurazakayashi/libNyaruko_Go/nyahttphandle"
	"github.com/tongdysoft/tongdySmallTools"
)

func sse(w http.ResponseWriter, req *http.Request, c chan []byte) {
	tongdySmallTools.PublicHandle(w, req)

	localeID := defaultLocaleID

	if req.Method == http.MethodOptions {
		c <- nyahttphandle.AlertInfoJson(w, localeID, 1001)
		return
	}

	req.ParseMultipartForm(32 << 20)
	fsid, ishsid := req.Form["sid"]        //会话ID
	fshowErr, ishshowErr := req.Form["se"] //是否显示错误信息
	fl, ishl := req.Form["l"]              //语言

	localeID = setLanguage(ishl, fl)
	showErr := false
	if ishshowErr && fshowErr[0] == "1" {
		showErr = true
	}

	if !ishsid {
		c <- nyahttphandle.AlertInfoJsonKV(w, localeID, 2040, "p", "sid")
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	runMsgKey := map[string]string{}
	times := 0
	for {
		select {
		case <-req.Context().Done(): // 客户端关闭连接
			fmt.Println("SSE: Client closed the connection. Exiting...")
			c <- nyahttphandle.AlertInfoJsonKV(w, 2, 10000, "", "Hello, SSE!")
			return
		case <-time.After(time.Duration(sendWaitTime) * time.Millisecond):
			keys := []string{}
			for k := range runMsgKey {
				if _, ok := messageSaveList[k]; !ok {
					delete(runMsgKey, k)
				}
			}
			for k := range messageSaveList {
				if _, ok := runMsgKey[k]; ok {
					continue
				}
				keys = append(keys, k)
				runMsgKey[k] = "run"
				fmt.Println(">>>>>>>>>>>>", k, runMsgKey)
			}
			// fmt.Println("SSE tick", fsid[0], messageSaveList, runMsgKey, keys)
			sort.Slice(keys, func(i, j int) bool {
				ki := strings.Split(keys[i], "_")[1]
				kj := strings.Split(keys[j], "_")[1]
				a, _ := strconv.Atoi(ki)
				b, _ := strconv.Atoi(kj)
				return a < b
			})

			getKeys := []string{}
			timeNano := []int64{}
			for i := 0; i < len(keys); i++ {
				key := keys[i]
				lastTimeStr := strings.Split(key, "_")[1]
				timestampInt, err := strconv.ParseInt(lastTimeStr, 10, 64)
				if err != nil {
					continue
				}
				tn := time.Now().UnixNano()
				tnOneSecondNano := tn - (sendWaitTime * 2 * 1e6)

				if times == 0 {
					nowHandleSessionID[fsid[0]] = "run"
					getKeys = append(getKeys, key)
					timeNano = append(timeNano, timestampInt)
				} else if tnOneSecondNano <= timestampInt && timestampInt <= tn {
					getKeys = append(getKeys, key)
					timeNano = append(timeNano, timestampInt)
				} else {
					continue
				}
			}

			if len(getKeys) <= 0 {
				times++
				delete(nowHandleSessionID, fsid[0])
				continue
			}

			data := []map[string]interface{}{}
			userIDs := []string{}
			msgs := []string{}
			createTime := []string{}
			delKeys := []string{}
			for i := 0; i < len(getKeys); i++ {
				key := getKeys[i]

				valData, ok := messageSaveList[key]
				if !ok {
					continue
				}

				timestamp := timeNano[i] / 1e6

				msg := map[string]interface{}{}

				msg["time"] = timestamp
				userID := valData[0]

				userMap, ok := userSaveMap[userID]
				if !ok {
					userMap["id"] = userID
					msg["user"] = userMap
				}

				msg["userID"] = userID
				msg["content"] = valData[1]

				if userID != "" {
					userIDs = append(userIDs, userID)
					msgs = append(msgs, valData[1])
					timeStr := time.Unix(0, timeNano[i]).In(cstSh).Format(timeFormat)
					createTime = append(createTime, timeStr)
					delKeys = append(delKeys, key)
				}

				data = append(data, msg)
			}

			dataStr, err := json.Marshal(data)
			if err == nil {
				fmt.Println("================")
				fmt.Println("dataStr", string(dataStr))
				fmt.Fprintf(w, "data:%s\n\n", string(dataStr))
				flusher.Flush()
				times++
			}

			reDelKeys := make(chan []string)

			go sseWriteSQL(fsid[0], userIDs, msgs, createTime, showErr, delKeys, reDelKeys)

			delKeys = <-reDelKeys
			for _, v := range delKeys {
				delete(runMsgKey, v)
			}
		}
	}
}

func sseHandleFunc(w http.ResponseWriter, req *http.Request) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	c := make(chan []byte)
	go sse(w, req, c)
	re := <-c
	wg.Done()
	w.Write([]byte(re))
	// fmt.Fprint(w, re)
	wg.Wait()
}

func sseWriteSQL(sessionID string, userID []string, msg []string, createTime []string, showErr bool, delKeys []string, reDelKeys chan []string) {
	nMi, err := mysqlIsRun(showErr)
	if err != nil {
		mysqlClose(nMi, true)
		reDelKeys <- []string{}
		return
	}
	// 将消息保存到数据库
	key := "`user_id`,`session_id`,`msg`,`creat_time`"
	cTime := ""
	values := ""
	for i := 0; i < len(userID); i++ {
		uID := userID[i]
		if uID == "" {
			uID = "1"
		}
		values += "('" + uID + "','" + sessionID + "','" + msg[i] + "','" + createTime[i] + "'),"
		cTime = createTime[i]
	}
	values = strings.TrimRight(values, ",")
	fmt.Println("delKeys", delKeys)

	inserts, err := nyaMList[nMi].AddRecord("message", key, "", values, nil)
	if err != nil {
		mysqlClose(nMi, true)
		reDelKeys <- []string{}
		return
	}
	msgID := strconv.Itoa(int(inserts))

	where := "`id`='" + sessionID + "'"
	update := fmt.Sprintf("`update_time`='%s',`update_id`='%s'", cTime, msgID)

	_, err = nyaMList[nMi].UpdataRecord("session", update, where, nil)
	mysqlClose(nMi, true)
	if err != nil {
		mysqlClose(nMi, true)
		reDelKeys <- []string{}
		return
	}

	delete(nowHandleSessionID, sessionID)

	fmt.Println("delKeys", delKeys)

	for _, v := range delKeys {
		delete(messageSaveList, v)
	}

	reDelKeys <- delKeys
}
