package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

var (
	lock1              sync.Mutex
	nowHandleSessionID map[string]string = map[string]string{} //当前正在处理的会话ID

	lock2           sync.Mutex
	messageSaveList map[string][]string = map[string][]string{} //消息列表

	lock3       sync.Mutex
	userSaveMap map[string]map[string]string = map[string]map[string]string{} //用户列表

	sendWaitTime int64 = 100 // 等待时间，单位毫秒
)
var (
	timeZoneInt     int    = 8               //时区
	timeZone        string = "Asia/Shanghai" //时区
	defaultLocaleID int    = 1               //默认语言

	tokenSaveTime int = 10 //token保存时间

	langs []string = []string{"en", "zhHans", "zhHant", "es"} //语言列表

	cstSh      *time.Location = nil                   //时区
	timeFormat string         = "2006-01-02 15:04:05" //时间格式

	qd  map[string]map[string]string       //查询结果
	err error                        = nil //错误信息
)

func main() {
	tn := time.Now()

	//获取设置
	err = getPublicVariable()
	if err != nil {
		fmt.Println("读取配置文件失败:", err)
		return
	}

	for i := 0; i < mysqlMaxLink; i++ {
		nyaMList = append(nyaMList, nil)
	}
	for i := 0; i < redisMaxLink; i++ {
		nyaRList = append(nyaRList, nil)
	}
	setupCloseHandler()

	cstSh, err = time.LoadLocation(timeZone)
	if err != nil {
		println("时区文件加载失败:", err)
		cstSh = time.FixedZone("CST", timeZoneInt*3600)
	}

	tStr := "[" + tn.In(cstSh).Format(timeFormat) + "]"
	println(tStr, "session-go v0.0.1")

	println("MySQL Link test")
	mI, err := mysqlIsRun(true)
	if err != nil {
		println("MySQL Link failed:", err)
		return
	}
	mysqlClose(mI, true)
	println("MySQL test link success")

	println("Redis Link test")
	rI, err := redisIsRun(redisSessionDB, true)
	if err != nil {
		println("Redis Link failed:", err)
		return
	}
	redisClose(rI, true)
	println("Redis test link success")

	// 处理Redis老数据
	go handleRedisData()

	println("监听端口:", listenandserve)
	println("初始化完成")
	println("====================")
	var handleFuncList []func(http.ResponseWriter, *http.Request)

	// 登录
	handleFuncList = append(handleFuncList, loginHandleFunc)

	// 发送信息
	handleFuncList = append(handleFuncList, sendMessageHandleFunc)

	// sse
	handleFuncList = append(handleFuncList, sseHandleFunc)

	if len(handleFuncList) != len(handleFuncKeyList) {
		println(len(handleFuncList), len(handleFuncKeyList))
		println("接口数量不匹配")
		return
	}

	http.HandleFunc(suburl, mainHandleFunc)
	println("接口列表:")
	println(suburl)
	for i := 0; i < len(handleFuncList); i++ {
		pattern := suburl + handleFuncKeyList[i]
		println(pattern)
		http.HandleFunc(pattern, handleFuncList[i])
	}

	err = http.ListenAndServe(":"+listenandserve, nil)
	if err != nil {
		println(err)
		return
	}
}

func mainHandleFunc(w http.ResponseWriter, r *http.Request) {
	println("mainHandleFunc")
	w.WriteHeader(404)
	w.Write([]byte{})
}
