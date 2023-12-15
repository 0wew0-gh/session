package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kagurazakayashi/libNyaruko_Go/nyahttphandle"
	"github.com/kagurazakayashi/libNyaruko_Go/nyaio"
)

func getPublicVariable() error {
	confStr, err := nyaio.FileRead("./setting.json")
	if err != nil {
		return err
	}

	var conf map[string]map[string]interface{}

	err = json.Unmarshal([]byte(confStr), &conf)
	if err != nil {
		return err
	}
	temp, ok := conf["dbSetting"]
	if ok {
		dbSettingConf := temp
		temp, ok := dbSettingConf["mysql"]
		if ok {
			byte, err := json.Marshal(temp.(map[string]interface{}))
			if err != nil {
				return err
			}
			mysqlconfig = string(byte)
		} else {
			return fmt.Errorf("缺少设置:'mysql'")
		}

		temp, ok = dbSettingConf["redis"]
		if ok {
			byte, err := json.Marshal(temp.(map[string]interface{}))
			if err != nil {
				return err
			}
			redisconfig = string(byte)
		} else {
			return fmt.Errorf("缺少设置:'redis'")
		}

		temp, ok = dbSettingConf["redis_max_db"]
		if ok {
			redisMaxDB = int(temp.(float64))
		} else {
			return fmt.Errorf("缺少设置:'redis_max_db'")
		}

		temp, ok = dbSettingConf["maxLinkNumber"]
		if ok {
			mysqlMaxLink = int(temp.(map[string]interface{})["mysql"].(float64))
			redisMaxLink = int(temp.(map[string]interface{})["redis"].(float64))
		} else {
			return fmt.Errorf("缺少设置:'maxLinkNumber'")
		}

		temp, ok = dbSettingConf["waitCount"]
		if ok {
			waitCount = int(temp.(float64))
		} else {
			println("缺少设置:'waitCount'")
		}

		temp, ok = dbSettingConf["waitTime"]
		if ok {
			waitTime = int(temp.(float64))
		} else {
			println("缺少设置:'waitTime'")
		}
	} else {
		return fmt.Errorf("缺少设置:'dbSetting'")
	}

	var config map[string]interface{}
	temp, ok = conf["config"]
	if ok {
		config = temp
	} else {
		return fmt.Errorf("缺少设置:'config'")
	}

	ctemp, ok := config["returnMessageFilePath"]
	if ok {
		nyahttphandle.AlertInfoTemplateLoad(ctemp.(string))
	} else {
		return fmt.Errorf("缺少设置:'returnMessageFilePath'")
	}

	ctemp, ok = config["timeZone"]
	if ok {
		tz := ctemp.([]interface{})
		if len(tz) != 2 {
			return fmt.Errorf("设置:'timeZone'格式错误")
		}
		timeZoneInt = int(ctemp.([]interface{})[0].(float64))
		timeZone = ctemp.([]interface{})[1].(string)
	} else {
		return fmt.Errorf("缺少设置:'timeZone'")
	}

	ctemp, ok = config["lang"]
	if ok {
		langs = []string{}
		for _, v := range ctemp.([]interface{}) {
			langs = append(langs, v.(string))
		}
	} else {
		return fmt.Errorf("缺少设置:'lang'")
	}

	ctemp, ok = config["suburl"]
	if ok {
		suburl = ctemp.(string)
	} else {
		return fmt.Errorf("缺少设置:'suburl'")
	}

	ctemp, ok = config["handleFuncKeyList"]
	if ok {
		for _, v := range ctemp.([]interface{}) {
			handleFuncKeyList = append(handleFuncKeyList, v.(string))
		}
	} else {
		return fmt.Errorf("缺少设置:'handleFuncKeyList'")
	}

	ctemp, ok = config["defaultLocaleID"]
	if ok {
		defaultLocaleID = int(ctemp.(float64))
	} else {
		println("缺少设置:'defaultLocaleID'")
	}

	ctemp, ok = config["redis_session_db"]
	if ok {
		redisSessionDB = int(ctemp.(float64))
	} else {
		return fmt.Errorf("缺少设置:'redis_session_db'")
	}

	ctemp, ok = config["redis_vcode_db"]
	if ok {
		redisVcodeDB = int(ctemp.(float64))
	} else {
		return fmt.Errorf("缺少设置:'redis_vcode_db'")
	}

	ctemp, ok = config["redis_login_db"]
	if ok {
		redisLoginDB = int(ctemp.(float64))
	} else {
		return fmt.Errorf("缺少设置:'redis_login_db'")
	}

	ctemp, ok = config["token_save_time"]
	if ok {
		tokenSaveTime = int(ctemp.(float64))
	} else {
		println("缺少设置:'token_save_time'")
	}

	return err
}

func setupCloseHandler() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("收到中止请求, 正在退出 ... ")
		fmt.Println("退出。")
		os.Exit(0)
	}()
}

func setLanguage(ishl bool, fl []string) int {
	if ishl {
		switch fl[0] {
		case "en", "1":
			return 1
		case "zh", "zh_cn", "zhHans", "chs", "2":
			return 2
		case "zh_tw", "zh_hk", "zhHant", "cht", "3":
			return 3
		case "es", "4":
			return 4
		}
	}
	return defaultLocaleID
}

// 判断参数是否为空
func ParameterIsNULL(str string) bool {
	if str == "" || str == "NULL" || str == "null" || str == "nil" || str == "undefined" {
		return true
	}
	return false
}

// ==========
//
//	组装where语句
//	where		string	"原where语句"
//	new		string	"需要加入where的语句"
//	delimiter	string	"连接符"
//	isApostrophe	bool	"是否需要加入单引号"
func assembleWhere(where string, new string, delimiter string, isApostrophe bool) string {
	if where != "" {
		where += delimiter
	}
	if isApostrophe {
		where += "'"
	}
	where += new
	if isApostrophe {
		where += "'"
	}
	return where
}
