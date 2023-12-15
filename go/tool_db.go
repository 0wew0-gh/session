package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/kagurazakayashi/libNyaruko_Go/nyacrypt"
	"github.com/kagurazakayashi/libNyaruko_Go/nyahttphandle"
	"github.com/kagurazakayashi/libNyaruko_Go/nyamysql"
	"github.com/kagurazakayashi/libNyaruko_Go/nyaredis"
)

var (
	mysqlconfig  string               = "{}" // mysql设置
	mysqlLink    int                  = 0    // mysql当前连接数
	mysqlMaxLink int                  = 10   // mysql最大连接数
	nyaMList     []*nyamysql.NyaMySQL = nil  // mysql对象数组
)

var (
	redisconfig  string               = "{}" // redis设置
	redisLink    int                  = 0    // redis当前连接数
	redisMaxLink int                  = 10   // redis最大连接数
	redisMaxDB   int                  = 16   // redis最大数据库数
	nyaRList     []*nyaredis.NyaRedis = nil  // redis对象数组
)

var (
	handleFuncKeyList []string = []string{} // 接口名称列表
	listenandserve    string   = "36041"    // 监听端口
	suburl            string   = "/"        // 子目录
	waitCount         int      = 10         // 等待次数
	waitTime          int      = 500        // 等待时间

	redisVcodeDB   int = 1 // redis验证码数据库
	redisLoginDB   int = 3 // redis登录数据库
	redisSessionDB int = 4 // redis session数据库
)

// ===================
//
//	连接MySQL数据库并放入连接池
//	waitCount	int	"等待次数"
//	waitTime	int	"每次等待时间，单位毫秒"
//	isShowPrint	bool	"是否显示连接信息"
//
//	return 1	int	"连接池中的位置"
//	return 2	error	"错误信息"
func mysqlIsRun(isShowPrint bool) (int, error) {
	if mysqlLink >= mysqlMaxLink {
		wc := 0
		for {
			if mysqlLink < mysqlMaxLink {
				break
			}
			if wc > waitCount {
				return -1, fmt.Errorf("MySQL connections are full")
			}
			wc += 1
			time.Sleep(time.Duration(waitTime) * time.Millisecond)
		}
	}
	// println("==========\r\nMySQL连接中...")
	nyaMS := nyamysql.New(mysqlconfig)
	if nyaMS.Error() != nil {
		if isShowPrint {
			println("MySQL DB Link error:", nyaMS.Error().Error())
		}
		return -1, err
	}
	ii := 0
	for i := 0; i < len(nyaMList); i++ {
		if nyaMList[i] == nil {
			ii = i
			break
		}
	}
	mysqlLink += 1
	nyaMList[ii] = nyaMS
	fmt.Println("MySQL connection successful!")
	fmt.Println("MySQL Number:", mysqlLink)
	return ii, nil
}

// ===================
//
//	关闭MySQL连接
//	i		int	"连接池中的位置"
//	isShowPrint	bool	"是否显示关闭信息"
func mysqlClose(i int, isShowPrint bool) {
	if i < 0 || i >= len(nyaMList) {
		fmt.Println("MySQL is Close!")
		return
	}
	if nyaMList[i] != nil {
		nyaMList[i].Close()
		nyaMList[i] = nil
		mysqlLink -= 1
		if mysqlLink < 0 {
			mysqlLink = 0
		}
		if isShowPrint {
			println("MySQL Close Connection! Current number of connections:", mysqlLink)
		}
	}
}

// ===============
//
//	连接Redis数据库并放入连接池
//	dbID		int	"数据库在配置中的位置"
//	WaitCount	int	"等待次数"
//	WaitTime	int	"每次等待时间，单位毫秒"
//	IsShowPrint	bool	"是否输出到控制台"
//
//	return 1	int	"连接池中的位置"
//	return 2	error	"错误信息"
func redisIsRun(dbID int, isShowPrint bool) (int, error) {
	if redisLink >= redisMaxLink {
		wc := 0
		for {
			if redisLink < redisMaxLink {
				break
			}
			if wc > waitCount {
				return -1, fmt.Errorf("MySQL connections are full")
			}
			wc += 1
			time.Sleep(time.Duration(waitTime) * time.Millisecond)
		}
	}
	nyaR := nyaredis.NewDB(redisconfig, dbID, redisMaxDB)
	if nyaR.Error() != nil {
		if isShowPrint {
			println("MySQL DB Link error:", nyaR.Error().Error())
		}
		return -1, err
	}
	ii := 0
	for i := 0; i < len(nyaRList); i++ {
		if nyaRList[i] == nil {
			ii = i
			break
		}
	}
	redisLink += 1
	nyaRList[ii] = nyaR
	return ii, nil
}

// ===============
//
//	关闭Redis连接
//	i		int	"连接池中的位置"
//	isShowPrint	bool	"是否显示关闭信息"
func redisClose(i int, isShowPrint bool) {
	if i < 0 || i >= len(nyaRList) {
		return
	}
	if nyaRList[i] != nil {
		nyaRList[i].Close()
		nyaRList[i] = nil
		redisLink -= 1
		if redisLink < 0 {
			redisLink = 0
		}
		if isShowPrint {
			println("Redis Close Connection! Current number of connections:", redisLink)
		}
	}
}

// 将用户信息存入redis并返回token
func setUserInfo(userInfo map[string]interface{}, isShowPrint bool) (string, int, error) {
	var (
		redisKey string
		err      error
	)
	// 生成token
	tn := time.Now().UnixNano()
	redisKey = nyacrypt.MD5String(strconv.Itoa(int(tn)), "")

	bytes, err := json.Marshal(userInfo)
	if err != nil {
		return "", 9999, err
	}
	// 将用户信息存入redis
	rI, err := redisIsRun(redisLoginDB, isShowPrint)
	if err != nil {
		return "", 9010, err
	}

	nyaRList[rI].SetString("ow_"+redisKey, string(bytes), nyaredis.Option_autoDelete(86400*tokenSaveTime))
	err = nyaRList[rI].Error()
	if err != nil {
		redisClose(rI, isShowPrint)
		fmt.Println("ow_"+redisKey, string(bytes))
		return "", 9011, err
	}
	redisClose(rI, isShowPrint)
	return redisKey, -1, nil
}

func backErrorMsg(w http.ResponseWriter, localeID int, errCode int, err interface{}, showErr bool, data interface{}) []byte {
	// 定义正则表达式
	re := regexp.MustCompile(`key '(.+)_UNIQUE'`)
	switch errType := err.(type) {
	case error:
		errStr := errType.Error()
		match := re.FindStringSubmatch(errStr)
		for _, vv := range match {
			fmt.Println(">>", vv)
		}
		if len(match) < 2 {
			break
		}
		errCode = 9008
		errStrList := strings.Split(match[1], ".")
		fmt.Println(errStrList)
		if len(errStrList) < 2 {
			break
		}
		err = errStrList[1]
	case []string:
		errStr := ""
		for i, v := range errType {
			match := re.FindStringSubmatch(v)
			for _, vv := range match {
				fmt.Println(">>", vv)
			}

			if len(match) < 2 {
				continue
			}
			errCode = 9008
			errStrList := strings.Split(match[1], ".")
			if len(errStrList) < 2 {
				continue
			}
			if i == 0 {
				errStr = ""
			}
			if errStr != "" {
				errStr += ","
			}
			errStr += errStrList[1]
		}
		if errStr != "" {
			err = errStr
		}
	}
	reData := map[string]interface{}{}
	reData["data"] = data
	if showErr {
		reData["err"] = err
		return nyahttphandle.AlertInfoJsonKV(w, localeID, errCode, "", reData)
	}
	if data != nil {
		return nyahttphandle.AlertInfoJsonKV(w, localeID, errCode, "", reData)
	}
	return nyahttphandle.AlertInfoJson(w, localeID, errCode)
}

// 验证token
func verifyToken(t string, isShowPrint bool) (map[string]interface{}, int, error) {
	rI, err := redisIsRun(redisLoginDB, isShowPrint)
	if err != nil {
		redisClose(rI, isShowPrint)
		return nil, 9010, err
	}

	userInfoStr := nyaRList[rI].GetString("ow_" + t)
	err = nyaRList[rI].Error()
	if err != nil {
		redisClose(rI, isShowPrint)
		if strings.Contains(err.Error(), "redis: nil") {
			return nil, 3900, err
		}
		return nil, 9011, err
	}
	redisClose(rI, isShowPrint)

	var userInfo map[string]interface{}
	err = json.Unmarshal([]byte(userInfoStr), &userInfo)
	if err != nil {
		return nil, 3901, err
	}

	return userInfo, -1, nil
}

// 对比验证码
func verifyVCode(user string, vcode string, isShowPrint bool) (bool, int, error) {
	rI, err := redisIsRun(redisVcodeDB, isShowPrint)
	if err != nil {
		redisClose(rI, isShowPrint)
		return false, 9010, err
	}
	temp := nyaRList[rI].GetString(user, nyaredis.Option_isDelete(true))
	err = nyaRList[rI].Error()
	redisClose(rI, isShowPrint)
	if err != nil {
		if err.Error() == "redis: nil" {
			return false, -1, nil
		}
		return false, 9011, err
	}
	if temp != vcode {
		return false, -1, nil
	}
	return true, -1, nil
}
