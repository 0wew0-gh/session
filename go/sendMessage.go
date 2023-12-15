package main

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/kagurazakayashi/libNyaruko_Go/nyahttphandle"
	"github.com/tongdysoft/tongdySmallTools"
)

func sendMessage(w http.ResponseWriter, req *http.Request, c chan []byte, mI chan int) {
	tongdySmallTools.PublicHandle(w, req)

	nMi := -1
	localeID := defaultLocaleID

	if req.Method == http.MethodOptions {
		mI <- nMi
		c <- nyahttphandle.AlertInfoJson(w, localeID, 1001)
		return
	} else if req.Method != http.MethodPost { // 检查是否为post请求
		// 返回 不是POST请求 的错误
		mI <- nMi
		c <- nyahttphandle.AlertInfoJson(w, localeID, 2001)
		return
	}

	req.ParseMultipartForm(32 << 20)
	ft, isht := req.Form["t"]              //token
	fm, ishm := req.Form["m"]              //用户名（邮件地址）
	fsessID, ishsessID := req.Form["sid"]  //会话ID，没有此项则新建会话
	fmsg, ishmsg := req.Form["msg"]        //消息内容
	fshowErr, ishshowErr := req.Form["se"] //是否显示错误信息
	fl, ishl := req.Form["l"]              //语言
	localeID = setLanguage(ishl, fl)
	userID := ""
	sessionID := ""
	showErr := false
	if ishshowErr && fshowErr[0] == "1" {
		showErr = true
	}
	if !ishmsg {
		mI <- nMi
		c <- nyahttphandle.AlertInfoJsonKV(w, localeID, 2040, "p", "msg")
		return
	} else if fmsg[0] == "" {
		mI <- nMi
		c <- nyahttphandle.AlertInfoJsonKV(w, localeID, 2041, "p", "msg cannot be empty")
		return
	}
	// if !ishm && !ishsessID {
	// 	mI <- nMi
	// 	c <- nyahttphandle.AlertInfoJsonKV(w, localeID, 2040, "p", "mail,uid,sid")
	// 	return
	// }

	userMap := map[string]string{}

	nMi, err = mysqlIsRun(showErr)
	if nMi == -1 {
		mI <- nMi
		c <- backErrorMsg(w, localeID, 9000, err, showErr, nil)
		return
	}

	// 生成查询用户的语句
	where := ""
	if isht {
		userInfo, errCode, err := verifyToken(ft[0], showErr)
		if err != nil {
			mI <- nMi
			c <- backErrorMsg(w, localeID, errCode, err, showErr, nil)
			return
		}
		temp, ok := userInfo["id"]
		if ok {
			switch tempType := temp.(type) {
			case string:
				userID = tempType
			case float64:
				userID = fmt.Sprint(tempType)
			}
		}
		temp, ok = userInfo["mail"]
		if ok {
			userMap["mail"] = temp.(string)
		}
		temp, ok = userInfo["nick"]
		if ok {
			userMap["nick"] = temp.(string)
		}
	} else if ishm || ishsessID {
		if showErr {
			fmt.Println(">>> 查询用户 <<<")
		}
		if ishm {
			where += "`mail`='" + fm[0] + "'"
		}
		qd, err = nyaMList[nMi].QueryData("`id`,`mail`,`nick`", "user", where, "", "", nil)
		if err != nil {
			mI <- nMi
			c <- backErrorMsg(w, localeID, 9001, err, showErr, nil)
			return
		}
		if len(qd) > 0 {
			if showErr {
				fmt.Println(">>> 如果查询有结果 <<<")
			}
			// 如果查询到多个说明 id 和 mail 不是同一个用户
			if len(qd) > 1 {
				mI <- nMi
				c <- nyahttphandle.AlertInfoJsonKV(w, localeID, 2042, "p", "mail/uid")
				return
			}
			fmt.Println(qd)
			fmt.Println(">>>", qd["0"], qd["0"]["id"])
			var dbi string

			for i := 0; i < len(qd); i++ {
				strI := strconv.Itoa(i)
				item := qd[strI]
				userID = item["id"]

				userMap["mail"] = item["mail"]
				userMap["nick"] = item["nick"]
			}
			fmt.Println("Decrypt:", userID, dbi, err)

		} else {
			if showErr {
				fmt.Println(">>> 如果查询 没有 结果 <<<")
			}
			if !ishm {
				mI <- nMi
				c <- nyahttphandle.AlertInfoJsonKV(w, localeID, 2040, "p", "mail/uid")
				return
			}
			if showErr {
				fmt.Println(">>> 添加新的用户 <<<")
			}
			inserts, err := nyaMList[nMi].AddRecord("user", "`mail`", "`"+fm[0]+"`", "", nil)
			if err != nil {
				mI <- nMi
				c <- backErrorMsg(w, localeID, 9002, err, showErr, nil)
				return
			}
			userID = strconv.Itoa(int(inserts))
		}
		if showErr {
			fmt.Println("用户ID", userID)

			fmt.Println(">>> 会话 <<<")
		}
	}

	// 如果有sid，则是发送给已有的会话
	// tn := time.Now().In(cstSh).Format(timeFormat)
	if !ishsessID {
		if showErr {
			fmt.Println(">>> 添加新的会话 <<<")
		}
		key := "`update_time`"
		val := "NOW()"
		if ishm {
			key += ",`mail`"
			val += ",'" + fm[0] + "'"
		}
		inserts, err := nyaMList[nMi].AddRecord("session", key, val, "", nil)
		if err != nil {
			mI <- nMi
			c <- backErrorMsg(w, localeID, 9002, err, showErr, nil)
			return
		}
		sessionID = strconv.Itoa(int(inserts))
	} else {
		sessionID = fsessID[0]
		where := "`id`='" + sessionID + "'"
		if showErr {
			fmt.Println(">>> 查询会话 <<<")
		}
		qd, err = nyaMList[nMi].QueryData("*", "session", where, "", "", nil)
		if err != nil {
			mI <- nMi
			c <- backErrorMsg(w, localeID, 9001, err, showErr, nil)
			return
		}
		if len(qd) > 0 {
			if showErr {
				fmt.Println(">>> 如果查询有结果 <<<")
			}
			if len(qd) > 1 {
				mI <- nMi
				c <- nyahttphandle.AlertInfoJsonKV(w, localeID, 2041, "p", "sid")
				return
			}
			sessionID = qd["0"]["id"]
		} else {
			if showErr {
				fmt.Println(">>> 如果查询 没有 结果 <<<")
			}
			mI <- nMi
			c <- nyahttphandle.AlertInfoJsonKV(w, localeID, 2041, "p", "sid")
			return
		}
	}
	if showErr {
		fmt.Println("会话ID", sessionID)
		fmt.Println("sid", fsessID, "msg", fmsg)
	}

	// rI, err := redisIsRun(redisSessionDB, showErr)
	// if err != nil {
	// 	redisClose(rI, true)
	// 	fmt.Println(nyaRList, redisLink)
	// 	mI <- nMi
	// 	c <- backErrorMsg(w, localeID, 9010, err, showErr, nil)
	// 	return
	// }
	tn := time.Now().UnixNano()

	redisKey := fmt.Sprintf("%s_%d", sessionID, tn)
	redisVal := []string{}
	redisVal = append(redisVal, userID)
	redisVal = append(redisVal, fmsg[0])
	// val, err := json.Marshal(redisVal)
	// if err != nil {
	// 	redisClose(rI, true)
	// 	mI <- nMi
	// 	c <- backErrorMsg(w, localeID, 9011, err, showErr, nil)
	// 	return
	// }
	// isDone := nyaRList[rI].SetString(redisKey, string(val))
	messageSaveList[redisKey] = redisVal

	if userID != "" {
		// redisUserStr := nyaRList[rI].GetString("u_" + userID)
		// fmt.Println(">>> get Redis User", redisUserStr)
		// if redisUserStr == "" {
		// 	fmt.Println(">>> set Redis User", userMap)
		// 	userStr, err := json.Marshal(userMap)
		// 	if err == nil {
		// 		nyaRList[rI].SetString("u_"+userID, string(userStr))
		// 	}
		// }
		// fmt.Println("redis", isDone, redisKey, redisVal)

		userSaveMap[userID] = userMap
	}
	// redisClose(rI, true)

	// // 将消息保存到数据库
	// key := "`user_id`,`session_id`,`msg`"
	// value := "'" + userID + "','" + sessionID + "','" + fmsg[0] + "'"
	// inserts, err := nyaMList[nMi].AddRecord("message", key, value, "", nil)
	// if err != nil {
	// 	mI <- nMi
	// 	c <- backErrorMsg(w, localeID, 9002, err, showErr, nil)
	// 	return
	// }
	// msgID := strconv.Itoa(int(inserts))

	// where = "`id`='" + sessionID + "'"
	// update := fmt.Sprintf("`update_time`=NOW(),`update_id`='%s'", msgID)

	// _, err = nyaMList[nMi].UpdataRecord("session", update, where, nil)
	// if err != nil {
	// 	mI <- nMi
	// 	c <- backErrorMsg(w, localeID, 9002, err, showErr, nil)
	// 	return
	// }
	data := map[string]string{}
	// data["messageID"] = msgID
	data["sessionID"] = sessionID
	data["userID"] = userID
	if uMail, ok := userMap["mail"]; ok {
		data["mail"] = uMail
	} else {
		data["mail"] = ""
	}
	time.Sleep(time.Duration(sendWaitTime) * time.Millisecond)
	mI <- nMi
	c <- nyahttphandle.AlertInfoJsonKV(w, localeID, 10000, "", data)
}

func sendMessageHandleFunc(w http.ResponseWriter, req *http.Request) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	c := make(chan []byte)
	cI := make(chan int)
	go sendMessage(w, req, c, cI)
	mI, re := <-cI, <-c
	wg.Done()
	mysqlClose(mI, true)
	w.Write([]byte(re))
	// fmt.Fprint(w, re)
	wg.Wait()
}
