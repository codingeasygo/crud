package crud

import (
	"fmt"
	"testing"

	"github.com/codingeasygo/util/converter"
	"github.com/codingeasygo/util/xsql"
)

type MaterialGroup struct {
	TID        int64     `json:"tid"`
	Sys        string    `json:"sys"`         //所属子系统
	UserID     int64     `json:"user_id"`     //所属用户id
	ParentID   int64     `json:"parent_id"`   //父ID，用于多级
	Title      *string   `json:"title"`       //组名
	OrderKey   string    `json:"order_key"`   //排序
	UpdateTime xsql.Time `json:"update_time"` //最后更新时间
	CreateTime xsql.Time `json:"create_time"` //创建时间
	Status     int       `json:"status"`      // 状态，查看MaterialGroupStatus*定义
}

func TestBuild(t *testing.T) {
	fmt.Println(BuildInsert("json", true, true, &MaterialGroup{TID: 1, Title: converter.StringPtr("")}))
}
