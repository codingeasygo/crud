package sqlx

import (
	"testing"

	"github.com/codingeasygo/util/xmap"
)

func TestMocker(t *testing.T) {
	Should(t, "ok", 1).OnlyLog(true).Call(func(trigger int) (res xmap.M, err error) {
		res = xmap.M{"a": 1}
		return
	})
	MockerSetCall("a", 1).OnlyLog(true).Should(t, "ok", 1).Call(func(trigger int) (res xmap.M, err error) {
		res = xmap.M{"a": 1}
		return
	})
	MockerPanicCall("a", 1).OnlyLog(true).Should(t, "ok", 1).Call(func(trigger int) (res xmap.M, err error) {
		res = xmap.M{"a": 1}
		return
	})
	MockerSetRangeCall("a", 1, 3).OnlyLog(true).Should(t, "ok", 1).Call(func(trigger int) (res xmap.M, err error) {
		res = xmap.M{"a": 1}
		return
	})
	MockerPanicRangeCall("a", 1, 3).OnlyLog(true).Should(t, "ok", 1).Call(func(trigger int) (res xmap.M, err error) {
		res = xmap.M{"a": 1}
		return
	})
	MockerMatchSetCall("a", "sql").OnlyLog(true).Should(t, "ok", 1).Call(func(trigger int) (res xmap.M, err error) {
		res = xmap.M{"a": 1}
		return
	})
	MockerMatchPanicCall("a", "sql").OnlyLog(true).Should(t, "ok", 1).Call(func(trigger int) (res xmap.M, err error) {
		res = xmap.M{"a": 1}
		return
	})
}
