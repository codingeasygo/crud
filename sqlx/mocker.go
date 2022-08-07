package sqlx

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"sync"
	"testing"

	"github.com/codingeasygo/util/converter"
	"github.com/codingeasygo/util/xmap"
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

var Log = log.New(os.Stderr, "    ", log.Llongfile)

type MockerCaller struct {
	Call      func(func(trigger int64) (res xmap.M, err error))
	shouldErr *testing.T
	shouldKey string
	shouldVal interface{}
}

func (m *MockerCaller) Should(t *testing.T, key string, v interface{}) *MockerCaller {
	m.shouldErr, m.shouldKey, m.shouldVal = t, key, v
	return m
}

func (m *MockerCaller) callError(err error) {
	if m.shouldErr == nil {
		panic(err)
	} else {
		Log.Output(5, err.Error())
		m.shouldErr.Fail()
	}
}

func (m *MockerCaller) validError(res xmap.M, err error) bool {
	if err != nil {
		m.callError(err)
		return false
	}
	return true
}

func (m *MockerCaller) validShould(res xmap.M, err error) bool {
	if len(m.shouldKey) < 1 {
		return true
	}
	val := res.Value(m.shouldKey)
	if m.shouldVal == nil || val == nil {
		if m.shouldVal == nil && val == nil {
			return true
		}
		m.callError(fmt.Errorf("res.%v(%v)!=%v, result is %v", m.shouldKey, val, m.shouldVal, converter.JSON(res)))
		return false
	}
	resultValue := reflect.ValueOf(val)
	if !resultValue.CanConvert(reflect.TypeOf(m.shouldVal)) {
		m.callError(fmt.Errorf("res.%v(%v)!=%v, result is %v", m.shouldKey, val, m.shouldVal, converter.JSON(res)))
		return false
	}
	targetValue := resultValue.Convert(reflect.TypeOf(m.shouldVal))
	if !reflect.DeepEqual(targetValue, m.shouldVal) {
		m.callError(fmt.Errorf("res.%v(%v)!=%v, result is %v", m.shouldKey, val, m.shouldVal, converter.JSON(res)))
		return false
	}
	return true
}

func (m *MockerCaller) validResult(res xmap.M, err error) {
	if !m.validError(res, err) {
		return
	}
	if !m.validShould(res, err) {
		return
	}
}

func Should(t *testing.T, key string, v interface{}) (caller *MockerCaller) {
	caller = &MockerCaller{}
	caller.Should(t, key, v).Call = func(call func(trigger int64) (res xmap.M, err error)) {
		res, err := call(0)
		caller.validResult(res, err)
	}
	return
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

func MockerSetCall(key string, triggers ...int64) (caller *MockerCaller) {
	caller = &MockerCaller{}
	caller.Call = func(call func(trigger int64) (res xmap.M, err error)) {
		for _, i := range triggers {
			MockerSet(key, i)
			res, err := call(i)
			MockerClear()
			caller.validResult(res, err)
		}
	}
	return
}

func MockerPanicCall(key string, triggers ...int64) (caller *MockerCaller) {
	caller = &MockerCaller{}
	caller.Call = func(call func(trigger int64) (res xmap.M, err error)) {
		for _, i := range triggers {
			MockerPanic(key, i)
			res, err := call(i)
			MockerClear()
			caller.validResult(res, err)
		}
	}
	return
}

func MockerMatchSetCall(key, match string) (caller *MockerCaller) {
	caller = &MockerCaller{}
	caller.Call = func(call func(trigger int64) (res xmap.M, err error)) {
		MockerMatchSet(key, match)
		res, err := call(0)
		MockerClear()
		caller.validResult(res, err)
	}
	return
}

func MockerMatchPanicCall(key, match string) (caller *MockerCaller) {
	caller = &MockerCaller{}
	caller.Call = func(call func(trigger int64) (res xmap.M, err error)) {
		MockerMatchPanic(key, match)
		res, err := call(0)
		MockerClear()
		caller.validResult(res, err)
	}
	return
}

func MockerSetRangeCall(key string, start, end int64) (caller *MockerCaller) {
	caller = &MockerCaller{}
	caller.Call = func(call func(trigger int64) (res xmap.M, err error)) {
		for i := start; i < end; i++ {
			MockerSet(key, i)
			res, err := call(0)
			MockerClear()
			caller.validResult(res, err)
		}
	}
	return
}

func MockerPanicRangeCall(key string, start, end int64) (caller *MockerCaller) {
	caller = &MockerCaller{}
	caller.Call = func(call func(trigger int64) (res xmap.M, err error)) {
		for i := start; i < end; i++ {
			MockerPanic(key, i)
			res, err := call(0)
			MockerClear()
			caller.validResult(res, err)
		}
	}
	return
}
