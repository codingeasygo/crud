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

type MockerCaller struct {
	Call func(func(trigger int64))
}

func MockerSet(key string, trigger int64) {
	mockerSet(key, "", false, trigger)
}

func MockerPanic(key string, trigger int64) {
	mockerSet(key, "", true, trigger)
}

func MockerMatchSet(key, match string) {
	mockerSet(key, match, false)
}

func MockerMatchPanic(key, match string) {
	mockerSet(key, match, true)
}

func MockerSetCall(key string, triggers ...int64) MockerCaller {
	return MockerCaller{
		Call: func(call func(trigger int64)) {
			for _, i := range triggers {
				MockerSet(key, i)
				call(i)
				MockerClear()
			}
		},
	}
}

func MockerMatchSetCall(key, match string) MockerCaller {
	return MockerCaller{
		Call: func(call func(trigger int64)) {
			MockerMatchSet(key, match)
			defer MockerClear()
			call(0)
		},
	}
}

func MockerMatchPanicCall(key, match string) MockerCaller {
	return MockerCaller{
		Call: func(call func(trigger int64)) {
			MockerMatchPanic(key, match)
			defer MockerClear()
			call(0)
		},
	}
}

func MockerPanicCall(key string, triggers ...int64) MockerCaller {
	return MockerCaller{
		Call: func(call func(trigger int64)) {
			for _, i := range triggers {
				MockerPanic(key, i)
				call(i)
				MockerClear()
			}
		},
	}
}

func MockerSetRangeCall(key string, start, end int64) MockerCaller {
	return MockerCaller{
		Call: func(call func(trigger int64)) {
			for i := start; i < end; i++ {
				MockerSet(key, i)
				call(i)
				MockerClear()
			}
		},
	}
}

func MockerPanicRangeCall(key string, start, end int64) MockerCaller {
	return MockerCaller{
		Call: func(call func(trigger int64)) {
			for i := start; i < end; i++ {
				MockerPanic(key, i)
				call(i)
				MockerClear()
			}
		},
	}
}
