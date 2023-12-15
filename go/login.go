package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/kagurazakayashi/libNyaruko_Go/nyahttphandle"
	"github.com/tongdysoft/tongdySmallTools"
)

func login(w http.ResponseWriter, req *http.Request, c chan []byte, mI chan int) {
	// fmt.Println("> loginHandleFunc")
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
	// 登录时，验证码与密码有一个即可
	// 注册时，验证码必须存在
	fuser, ishuser := req.Form["user"]     //邮箱
	fpw, ishpw := req.Form["pw"]           //密码
	fvcode, ishvcode := req.Form["vcode"]  //验证码
	fnick, ishnick := req.Form["n"]        //昵称
	fowb, ishowb := req.Form["owb"]        //是否是后端，如果是官网前端则只返回是否不是普通用户，如果是官网后台则返回详细权限
	fp, ishp := req.Form["p"]              //权限
	flang, ishlang := req.Form["lang"]     //创建用户时的语言
	fshowErr, ishshowErr := req.Form["se"] //是否显示错误信息
	fl, ishl := req.Form["l"]              //语言
	localeID = setLanguage(ishl, fl)
	showErr := false
	if ishshowErr && fshowErr[0] == "1" {
		showErr = true
	}
	var code2040 []string
	if !ishuser {
		code2040 = append(code2040, "html")
	}
	if len(code2040) > 0 {
		code2040str := strings.Join(code2040, ",")
		mI <- nMi
		c <- nyahttphandle.AlertInfoJsonKV(w, localeID, 2040, "p", code2040str)
		return
	}

	isTrueCode := false
	where := "`mail`='" + fuser[0] + "' AND `del_time` IS NULL"
	// 验证码是否正确
	if ishvcode {
		var errCode int = -1
		isTrueCode, errCode, err = verifyVCode(fuser[0], fvcode[0], showErr)
		if errCode > 0 {
			mI <- nMi
			c <- backErrorMsg(w, localeID, errCode, err, showErr, nil)
			return
		}
		if !isTrueCode {
			mI <- nMi
			c <- nyahttphandle.AlertInfoJson(w, localeID, 4013)
			return
		}
	} else if ishpw {
		where += " AND `password`='" + fpw[0] + "'"
	}

	nMi, err = mysqlIsRun(showErr)
	if nMi == -1 {
		mI <- nMi
		c <- backErrorMsg(w, localeID, 9000, err, showErr, nil)
		return
	}

	// 根据邮箱（和密码）查询用户
	qd, err = nyaMList[nMi].QueryData("*", "user", where, "", "", nil)
	if err != nil {
		mI <- nMi
		c <- backErrorMsg(w, localeID, 9001, err, showErr, nil)
		return
	}
	// 如果查询到了用户
	if len(qd) > 0 {
		// 生成用户信息
		var (
			userInfo map[string]interface{} = map[string]interface{}{}
			ok       bool                   = false
		)
		userInfo["id"], ok = qd["0"]["id"]
		if !ok {
			mI <- nMi
			c <- nyahttphandle.AlertInfoJson(w, localeID, 4007)
			return
		}
		userInfo["mail"] = qd["0"]["mail"]
		if !ok {
			mI <- nMi
			c <- nyahttphandle.AlertInfoJson(w, localeID, 4007)
			return
		}
		userInfo["nick"] = qd["0"]["nick"]
		if !ok {
			mI <- nMi
			c <- nyahttphandle.AlertInfoJson(w, localeID, 4007)
			return
		}
		var permission interface{}
		permission, ok = qd["0"]["permission"]
		if !ok {
			mI <- nMi
			c <- nyahttphandle.AlertInfoJson(w, localeID, 4007)
			return
		}
		if ishowb && fowb[0] == "1" {
			userInfo["permission"] = permission
		} else {
			isSuperAdmin := true
			if permission == "99" {
				isSuperAdmin = false
			}
			userInfo["permission"] = isSuperAdmin
		}
		userInfo["language"] = qd["0"]["language"]

		redisKey, errCode, err := setUserInfo(userInfo, showErr)
		if err != nil {
			mI <- nMi
			c <- backErrorMsg(w, localeID, errCode, err, showErr, nil)
			return
		}
		data := map[string]interface{}{}
		data["token"] = redisKey
		data["permission"] = userInfo["permission"]

		msgCode := 10000
		if ishnick {
			msgCode = 10003
		}
		mI <- nMi
		c <- nyahttphandle.AlertInfoJsonKV(w, localeID, msgCode, "", data)
		return
	}
	// 如果没有查询到用户
	if ishpw {
		// 如果有密码，则只根据用户名查询是否有用户
		where = "`mail`='" + fuser[0] + "' AND `del_time` IS NULL"
		qd, err = nyaMList[nMi].QueryData("*", "user", where, "", "", nil)
		if err != nil {
			mI <- nMi
			c <- backErrorMsg(w, localeID, 9001, err, showErr, nil)
			return
		}
		// 如果查询到了用户，则返回密码错误
		if len(qd) > 0 {
			mI <- nMi
			c <- nyahttphandle.AlertInfoJson(w, localeID, 4012)
			return
		}
	}
	//创建用户必须有验证码
	if !isTrueCode {
		if ishvcode {
			mI <- nMi
			c <- nyahttphandle.AlertInfoJson(w, localeID, 4013)
			return
		}
		mI <- nMi
		c <- nyahttphandle.AlertInfoJson(w, localeID, 3999)
		return
	}

	// 创建账号
	key := "`mail`,`password`"
	value := "'" + fuser[0] + "','" + fpw[0] + "'"
	if ishnick {
		key = assembleWhere(key, "`nick`", ",", false)
		if ParameterIsNULL(fnick[0]) {
			value = assembleWhere(value, "NULL", ",", false)
		} else {
			value = assembleWhere(value, fnick[0], ",", true)
		}
	}
	if ishp && !ParameterIsNULL(fp[0]) {
		key = assembleWhere(key, "`permission`", ",", false)
		value = assembleWhere(value, fp[0], ",", true)
	}
	if ishlang && !ParameterIsNULL(flang[0]) {
		key = assembleWhere(key, "`language`", ",", false)
		value = assembleWhere(value, flang[0], ",", true)
	}

	fmt.Println(">>>>>>>>>>>>>>>>")
	fmt.Println(req.Form)
	fmt.Println(ishnick)
	fmt.Println(ishp, !ParameterIsNULL(fp[0]))
	fmt.Println(ishlang, !ParameterIsNULL(flang[0]))
	fmt.Println(">>>>>>>>>>>>>>>>")

	inserts, err := nyaMList[nMi].AddRecord("user", key, value, "", nil)
	if err != nil {
		fmt.Println(err)
		mI <- nMi
		c <- backErrorMsg(w, localeID, 9001, err, showErr, nil)
		return
	}

	if showErr {
		fmt.Println("新建用户 ID:", inserts)
	}

	// 处理用户
	userInfo := map[string]interface{}{}
	userInfo["id"] = inserts
	if ishuser {
		userInfo["mail"] = fuser[0]
	}
	if ishnick {
		userInfo["nick"] = fnick[0]
	} else {
		userInfo["nick"] = ""
	}
	userInfo["permission"] = false
	userInfo["language"] = localeID
	// 将用户信息保存到redis
	redisKey, errCode, err := setUserInfo(userInfo, showErr)
	if err != nil {
		mI <- nMi
		c <- backErrorMsg(w, localeID, errCode, err, showErr, nil)
		return
	}
	data := map[string]interface{}{}
	data["token"] = redisKey
	data["permission"] = false
	mI <- nMi
	c <- nyahttphandle.AlertInfoJsonKV(w, localeID, 10000, "", data)
}

func loginHandleFunc(w http.ResponseWriter, req *http.Request) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	cI := make(chan int)
	c := make(chan []byte)
	go login(w, req, c, cI)
	mI, re := <-cI, <-c
	wg.Done()
	mysqlClose(mI, true)
	w.Write([]byte(re))
	// fmt.Fprint(w, re)
	wg.Wait()
}
