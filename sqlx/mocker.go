package sqlx

import (
	"fmt"
	"regexp"
	"sync"
)

var ErrMock = fmt.Errorf("mock error")
var Verbose = false

var mocking = false
var mockPanic = false
var mockTrigger = map[string][]int64{}
var mockMatch = map[string]*regexp.Regexp{}
var mockRunned = map[string]int64{}
var mockRunnedLck = sync.RWMutex{}

func mockerCheck(key, sql string) (err error) {
	if mocking {
		mockRunnedLck.Lock()
		mockRunned[key]++
		trigger := mockTrigger[key]
		runned := mockRunned[key]
		if trigger != nil && (trigger[0] < 0 || (trigger[0] <= runned && runned <= trigger[1])) {
			err = ErrMock
		}
		match := mockMatch[key]
		if match != nil && match.MatchString(sql) {
			err = ErrMock
		}
		if Verbose {
			fmt.Printf("Mocking %v trigger:%v,runned:%v,err:%v,sql:\n%v\n", key, mockTrigger[key], mockRunned[key], err, sql)
		}
		mockRunnedLck.Unlock()
		if mockPanic && err != nil {
			panic(err)
		}
	}
	return
}

func MockerStart() {
	mocking = true
}

func MockerStop() {
	MockerClear()
	mocking = false
}

func MockerClear() {
	mockRunnedLck.Lock()
	mockTrigger = map[string][]int64{}
	mockMatch = map[string]*regexp.Regexp{}
	mockRunned = map[string]int64{}
	mockPanic = false
	mockRunnedLck.Unlock()
}

func mockerSet(key, match string, isPanice bool, triggers ...int64) {
	mockRunnedLck.Lock()
	defer mockRunnedLck.Unlock()
	if len(match) > 0 {
		mockMatch[key] = regexp.MustCompile(match)
	} else {
		if len(triggers) == 1 {
			mockTrigger[key] = []int64{triggers[0], triggers[0]}
		} else if len(triggers) > 1 {
			mockTrigger[key] = triggers
		} else {
			panic("trigger is required")
		}
	}
	mockPanic = isPanice
}

func MockerSet(key string, triggers ...int64) {
	mockerSet(key, "", false, triggers...)
}

func MockerPanic(key string, triggers ...int64) {
	mockerSet(key, "", true, triggers...)
}

func MockerMatchSet(key, match string) {
	mockerSet(key, match, false)
}

func MockerMatchPanic(key, match string) {
	mockerSet(key, match, true)
}
