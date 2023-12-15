package main

import (
	"fmt"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

func handleRedisData() {
	fmt.Println("> handleRedisData()")
	var timeInterval int64 = 30

	times := 0
	for {
		if times%10 == 0 {
			runtime.GC()
		}
		time.Sleep(time.Duration(timeInterval) * time.Second)
		fmt.Println("> handleRedisData:", times, nowHandleSessionID)

		keys := []string{}

		for k := range messageSaveList {
			keys = append(keys, k)
		}

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
			temp := strings.Split(key, "_")

			sid := temp[0]
			if _, ok := nowHandleSessionID[sid]; ok {
				continue
			}

			lastTimeStr := temp[1]
			timestampInt, err := strconv.ParseInt(lastTimeStr, 10, 64)
			if err != nil {
				continue
			}
			tn := time.Now()
			tnOneSecondNano := tn.UnixNano() - timeInterval*1e9
			if tnOneSecondNano >= timestampInt {
				getKeys = append(getKeys, key)
				timeNano = append(timeNano, timestampInt)
			} else {
				continue
			}
		}

		if len(getKeys) <= 0 {
			times++
			continue
		}

		sessionIDs := []string{}
		userIDs := [][]string{}
		msgs := [][]string{}
		createTime := [][]string{}
		delKeys := [][]string{}
		for i := 0; i < len(getKeys); i++ {
			key := getKeys[i]
			valData, ok := messageSaveList[key]
			if !ok {
				continue
			}

			sessionID := strings.Split(key, "_")[0]
			timeStr := time.Unix(0, timeNano[i]).In(cstSh).Format(timeFormat)

			isAdd := false
			for i := 0; i < len(sessionIDs); i++ {
				if sessionIDs[i] == sessionID {
					userIDs[i] = append(userIDs[i], valData[0])
					msgs[i] = append(msgs[i], valData[1])

					createTime[i] = append(createTime[i], timeStr)

					delKeys[i] = append(delKeys[i], key)
					isAdd = true
				}
			}
			if !isAdd {
				sessionIDs = append(sessionIDs, sessionID)
				userIDs = append(userIDs, []string{valData[0]})
				msgs = append(msgs, []string{valData[1]})

				createTime = append(createTime, []string{timeStr})

				delKeys = append(delKeys, []string{key})
			}
		}
		for i := 0; i < len(sessionIDs); i++ {
			go sseWriteSQL(sessionIDs[i], userIDs[i], msgs[i], createTime[i], false, delKeys[i], nil)
		}
		times++
	}
}
