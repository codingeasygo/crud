package pgx

import (
	"fmt"
	"reflect"
	"regexp"
	"sync"
	"testing"

	"github.com/codingeasygo/util/xmap"
)

var ErrMock = fmt.Errorf("mock error")
var Verbose = false

var mocking = false
var mockPanic = false
var mockTrigger = map[string][]int{}
var mockMatch = map[string]*regexp.Regexp{}
var mockRunned = map[string]int{}
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
	mockTrigger = map[string][]int{}
	mockMatch = map[string]*regexp.Regexp{}
	mockRunned = map[string]int{}
	mockPanic = false
	mockRunnedLck.Unlock()
}

func mockerSet(key, match string, isPanice bool, triggers ...int) {
	mockRunnedLck.Lock()
	defer mockRunnedLck.Unlock()
	if len(match) > 0 {
		mockMatch[key] = regexp.MustCompile(match)
	} else {
		if len(triggers) == 1 {
			mockTrigger[key] = []int{triggers[0], triggers[0]}
		} else if len(triggers) > 1 {
			mockTrigger[key] = triggers
		} else {
			panic("trigger is required")
		}
	}
	mockPanic = isPanice
}

type MockerCaller struct {
	Call     func(func(trigger int) (res xmap.M, err error)) xmap.M
	Shoulder xmap.Shoulder
}

func (m *MockerCaller) Should(t *testing.T, args ...interface{}) *MockerCaller {
	m.Shoulder.Should(t, args...)
	return m
}

func (m *MockerCaller) ShouldError(t *testing.T) *MockerCaller {
	m.Shoulder.ShouldError(t)
	return m
}

func (m *MockerCaller) OnlyLog(only bool) *MockerCaller {
	m.Shoulder.OnlyLog(only)
	return m
}

func Should(t *testing.T, key string, v interface{}) (caller *MockerCaller) {
	caller = &MockerCaller{}
	caller.Should(t, key, v).Call = func(call func(trigger int) (res xmap.M, err error)) xmap.M {
		res, err := call(0)
		caller.Shoulder.Valid(4, res, err)
		return res
	}
	return
}

func rangeArgs(args []interface{}, call func(key string, trigger int)) {
	triggerAll := map[string][]int{}
	triggerKeys := []string{}
	triggerAdd := false
	for i, arg := range args {
		switch arg := arg.(type) {
		case string:
			if triggerAdd {
				triggerKeys = []string{}
			}
			triggerAdd = false
			triggerKeys = append(triggerKeys, arg)
		case int:
			triggerAdd = true
			for _, key := range triggerKeys {
				triggerAll[key] = append(triggerAll[key], arg)
			}
		default:
			panic(fmt.Sprintf("args[%v] is %v and not supported", i, reflect.TypeOf(arg)))
		}
	}
	for key, triggers := range triggerAll {
		for _, trigger := range triggers {
			call(key, trigger)
		}
	}
}

func MockerSet(key string, trigger int) {
	mockerSet(key, "", false, trigger)
}

func MockerPanic(key string, trigger int) {
	mockerSet(key, "", true, trigger)
}

func MockerMatchSet(key, match string) {
	mockerSet(key, match, false)
}

func MockerMatchPanic(key, match string) {
	mockerSet(key, match, true)
}

func MockerSetCall(args ...interface{}) (caller *MockerCaller) {
	caller = &MockerCaller{}
	caller.Call = func(call func(trigger int) (res xmap.M, err error)) xmap.M {
		rangeArgs(args, func(key string, i int) {
			MockerSet(key, i)
			res, err := call(i)
			MockerClear()
			caller.Shoulder.Valid(6, res, err)
		})
		return nil
	}
	return
}

func MockerPanicCall(args ...interface{}) (caller *MockerCaller) {
	caller = &MockerCaller{}
	caller.Call = func(call func(trigger int) (res xmap.M, err error)) xmap.M {
		rangeArgs(args, func(key string, i int) {
			MockerPanic(key, i)
			res, err := call(i)
			MockerClear()
			caller.Shoulder.Valid(6, res, err)
		})
		return nil
	}
	return
}

func MockerMatchSetCall(key, match string) (caller *MockerCaller) {
	caller = &MockerCaller{}
	caller.Call = func(call func(trigger int) (res xmap.M, err error)) xmap.M {
		MockerMatchSet(key, match)
		res, err := call(0)
		MockerClear()
		caller.Shoulder.Valid(4, res, err)
		return res
	}
	return
}

func MockerMatchPanicCall(key, match string) (caller *MockerCaller) {
	caller = &MockerCaller{}
	caller.Call = func(call func(trigger int) (res xmap.M, err error)) xmap.M {
		MockerMatchPanic(key, match)
		res, err := call(0)
		MockerClear()
		caller.Shoulder.Valid(4, res, err)
		return res
	}
	return
}

func MockerSetRangeCall(key string, start, end int) (caller *MockerCaller) {
	caller = &MockerCaller{}
	caller.Call = func(call func(trigger int) (res xmap.M, err error)) xmap.M {
		for i := start; i < end; i++ {
			MockerSet(key, i)
			res, err := call(0)
			MockerClear()
			caller.Shoulder.Valid(4, res, err)
		}
		return nil
	}
	return
}

func MockerPanicRangeCall(key string, start, end int) (caller *MockerCaller) {
	caller = &MockerCaller{}
	caller.Call = func(call func(trigger int) (res xmap.M, err error)) xmap.M {
		for i := start; i < end; i++ {
			MockerPanic(key, i)
			res, err := call(0)
			MockerClear()
			caller.Shoulder.Valid(4, res, err)
		}
		return nil
	}
	return
}
