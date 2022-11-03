package pgx

import (
	"fmt"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"sync"
	"testing"

	"github.com/codingeasygo/util/xhttp"
	"github.com/codingeasygo/util/xmap"
)

var ErrMock = fmt.Errorf("mock error")
var Verbose = false
var Client = xhttp.Shared

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

type MockerCaller struct {
	Call     func(func(trigger int) (res xmap.M, err error)) xmap.M
	calld    func(int, func(trigger int) (res xmap.M, err error)) xmap.M
	Client   *xhttp.Client
	Shoulder xmap.Shoulder
}

func NewMockerCaller() (caller *MockerCaller) {
	caller = &MockerCaller{Client: Client}
	caller.Call = func(c func(trigger int) (xmap.M, error)) xmap.M { return caller.calld(1, c) }
	return
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

//GetMap will get map from remote
func (m *MockerCaller) GetMap(format string, args ...interface{}) (data xmap.M, err error) {
	m.calld(1, func(trigger int) (xmap.M, error) {
		data, err = m.Client.GetMap(format, args...)
		return data, err
	})
	return
}

//GetHeaderMap will get map from remote
func (m *MockerCaller) GetHeaderMap(header xmap.M, format string, args ...interface{}) (data xmap.M, res *http.Response, err error) {
	m.calld(1, func(trigger int) (xmap.M, error) {
		data, res, err = m.Client.GetHeaderMap(header, format, args...)
		return data, err
	})
	return
}

//PostMap will get map from remote
func (m *MockerCaller) PostMap(body io.Reader, format string, args ...interface{}) (data xmap.M, err error) {
	m.calld(1, func(trigger int) (xmap.M, error) {
		data, err = m.Client.PostMap(body, format, args...)
		return data, err
	})
	return
}

//PostTypeMap will get map from remote
func (m *MockerCaller) PostTypeMap(contentType string, body io.Reader, format string, args ...interface{}) (data xmap.M, err error) {
	m.calld(1, func(trigger int) (xmap.M, error) {
		data, err = m.Client.PostTypeMap(contentType, body, format, args...)
		return data, err
	})
	return
}

//PostHeaderMap will get map from remote
func (m *MockerCaller) PostHeaderMap(header xmap.M, body io.Reader, format string, args ...interface{}) (data xmap.M, res *http.Response, err error) {
	m.calld(1, func(trigger int) (xmap.M, error) {
		data, res, err = m.Client.PostHeaderMap(header, body, format, args...)
		return data, err
	})
	return
}

//PostJSONMap will get map from remote
func (m *MockerCaller) PostJSONMap(body interface{}, format string, args ...interface{}) (data xmap.M, err error) {
	m.calld(1, func(trigger int) (xmap.M, error) {
		data, err = m.Client.PostJSONMap(body, format, args...)
		return data, err
	})
	return
}

//MethodBytes will do http request, read reponse and parse to map
func (m *MockerCaller) MethodMap(method string, header xmap.M, body io.Reader, format string, args ...interface{}) (data xmap.M, res *http.Response, err error) {
	m.calld(1, func(trigger int) (xmap.M, error) {
		data, res, err = m.Client.MethodMap(method, header, body, format, args...)
		return data, err
	})
	return
}

//PostFormMap will get map from remote
func (m *MockerCaller) PostFormMap(form xmap.M, format string, args ...interface{}) (data xmap.M, err error) {
	m.calld(1, func(trigger int) (xmap.M, error) {
		data, err = m.Client.PostFormMap(form, format, args...)
		return data, err
	})
	return
}

//PostMultipartMap will get map from remote
func (m *MockerCaller) PostMultipartMap(header, fields xmap.M, format string, args ...interface{}) (data xmap.M, err error) {
	m.calld(1, func(trigger int) (xmap.M, error) {
		data, err = m.Client.PostMultipartMap(header, fields, format, args...)
		return data, err
	})
	return
}

//UploadMap will get map from remote
func (m *MockerCaller) UploadMap(fields xmap.M, filekey, filename, format string, args ...interface{}) (data xmap.M, err error) {
	m.calld(1, func(trigger int) (xmap.M, error) {
		data, err = m.Client.UploadMap(fields, filekey, filename, format, args...)
		return data, err
	})
	return
}

func Should(t *testing.T, args ...interface{}) (caller *MockerCaller) {
	caller = NewMockerCaller()
	caller.calld = func(depth int, call func(trigger int) (res xmap.M, err error)) xmap.M {
		res, err := call(0)
		caller.Shoulder.Valid(depth+3, res, err)
		return res
	}
	return caller.Should(t, args...)
}

func ShouldError(t *testing.T) (caller *MockerCaller) {
	caller = NewMockerCaller()
	caller.calld = func(depth int, call func(trigger int) (res xmap.M, err error)) xmap.M {
		res, err := call(0)
		caller.Shoulder.Valid(depth+3, res, err)
		return res
	}
	return caller.ShouldError(t)
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

func MockerSetCall(args ...interface{}) (caller *MockerCaller) {
	caller = NewMockerCaller()
	caller.calld = func(depth int, call func(trigger int) (res xmap.M, err error)) xmap.M {
		rangeArgs(args, func(key string, i int) {
			MockerSet(key, i)
			res, err := call(i)
			MockerClear()
			caller.Shoulder.Valid(depth+5, res, err)
		})
		return nil
	}
	return
}

func MockerPanicCall(args ...interface{}) (caller *MockerCaller) {
	caller = NewMockerCaller()
	caller.calld = func(depth int, call func(trigger int) (res xmap.M, err error)) xmap.M {
		rangeArgs(args, func(key string, i int) {
			MockerPanic(key, i)
			res, err := call(i)
			MockerClear()
			caller.Shoulder.Valid(depth+5, res, err)
		})
		return nil
	}
	return
}

func MockerMatchSetCall(key, match string) (caller *MockerCaller) {
	caller = NewMockerCaller()
	caller.calld = func(depth int, call func(trigger int) (res xmap.M, err error)) xmap.M {
		MockerMatchSet(key, match)
		res, err := call(0)
		MockerClear()
		caller.Shoulder.Valid(depth+3, res, err)
		return res
	}
	return
}

func MockerMatchPanicCall(key, match string) (caller *MockerCaller) {
	caller = NewMockerCaller()
	caller.calld = func(depth int, call func(trigger int) (res xmap.M, err error)) xmap.M {
		MockerMatchPanic(key, match)
		res, err := call(0)
		MockerClear()
		caller.Shoulder.Valid(depth+3, res, err)
		return res
	}
	return
}

func MockerSetRangeCall(key string, start, end int) (caller *MockerCaller) {
	caller = NewMockerCaller()
	caller.calld = func(depth int, call func(trigger int) (res xmap.M, err error)) xmap.M {
		for i := start; i < end; i++ {
			MockerSet(key, i)
			res, err := call(0)
			MockerClear()
			caller.Shoulder.Valid(depth+3, res, err)
		}
		return nil
	}
	return
}

func MockerPanicRangeCall(key string, start, end int) (caller *MockerCaller) {
	caller = NewMockerCaller()
	caller.calld = func(depth int, call func(trigger int) (res xmap.M, err error)) xmap.M {
		for i := start; i < end; i++ {
			MockerPanic(key, i)
			res, err := call(0)
			MockerClear()
			caller.Shoulder.Valid(depth+3, res, err)
		}
		return nil
	}
	return
}
