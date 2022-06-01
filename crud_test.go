package crud

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestJoin(t *testing.T) {
	if Join([]int{}, ",", "", "") != "" {
		t.Error("error")
		return
	}
	if Join([]int{1, 2, 3}, ",", "", "") != "1,2,3" {
		t.Error("error")
		return
	}
	ival1, ival2, ival3 := 1, 2, 3
	if Join([]*int{&ival1, &ival2, &ival3}, ",", "", "") != "1,2,3" {
		t.Error("error")
		return
	}
	sval1, sval2, sval3 := "1", "2", "3"
	if Join([]*string{&sval1, &sval2, &sval3}, ",", "", "") != "1,2,3" {
		t.Error("error")
		return
	}
	func() {
		defer func() {
			recover()
		}()
		Join("xx", ",", "", "")
	}()
}

type PoolQueryer struct {
	QueryErr error
	ScanErr  error
}

func (t *PoolQueryer) Query(sql string, args ...interface{}) (rows Rows, err error) {
	title := "test"
	rows = &TestRows{
		Err:   t.ScanErr,
		Index: -1,
		Values: [][]interface{}{
			ScanArgs(&Simple{
				TID:        1,
				UserID:     100,
				Type:       "test",
				Title:      &title,
				Image:      &title,
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
				Status:     SimpleStatusNormal,
			}, false, false, "", ""),
			ScanArgs(&Simple{
				TID:        2,
				UserID:     0,
				Type:       "test",
				Title:      &title,
				Image:      &title,
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
				Status:     SimpleStatusNormal,
			}, false, false, "", ""),
			ScanArgs(&Simple{
				TID:        2,
				UserID:     0,
				Type:       "test",
				Title:      &title,
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
				Status:     SimpleStatusNormal,
			}, false, false, "", ""),
		},
	}
	err = t.QueryErr
	return
}

type TestRows struct {
	Values [][]interface{}
	Index  int
	Err    error
}

func (t *TestRows) Scan(dests ...interface{}) (err error) {
	err = t.Err
	values := t.Values[t.Index]
	for i, dest := range dests {
		destValue := reflect.Indirect(reflect.ValueOf(dest))
		destValue.Set(reflect.Indirect(reflect.ValueOf(values[i])))
	}
	return
}

func (t *TestRows) Next() bool {
	t.Index++
	return t.Index < len(t.Values)
}

func (t *TestRows) Close() {

}

const (
	SimpleTypeMaterial  = "material"
	SimpleTypeProduct   = "product"
	SimpleStatusNormal  = 100
	SimpleStatusRemoved = -1
)

var (
	SimpleTypeAll   = []string{SimpleTypeMaterial, SimpleTypeProduct}
	SimpleStatusAll = []int{SimpleStatusNormal, SimpleStatusRemoved}
)

type Simple struct {
	TID        int64     `json:"tid"`
	UserID     int64     `json:"user_id"`
	Type       string    `json:"type"`
	Title      *string   `json:"title"`
	Image      *string   `json:"image"`
	Data       string    `json:"data" conv:"::text"`
	UpdateTime time.Time `json:"update_time"`
	CreateTime time.Time `json:"create_time"`
	Status     int       `json:"status"`
	XX         string    `json:"-"`
}

//AddSimple will add material group to database
func AddSimple(simple *Simple) (err error) {
	simple.CreateTime = time.Now()
	simple.UpdateTime = time.Now()
	insertSQL, insertArg := InsertAllSQL(simple, "crud_simple", "returning tid")
	fmt.Printf("insert\n")
	fmt.Printf("   --->%v\n", insertSQL)
	fmt.Printf("   --->%v\n", insertArg)
	insertSQL1, insertArg1 := InsertSQL(simple, true, true, "", "update_time,create_time,status", "crud_simple", "returning tid")
	fmt.Printf("insert\n")
	fmt.Printf("   --->%v\n", insertSQL1)
	fmt.Printf("   --->%v\n", insertArg1)
	if strings.Contains(insertSQL1, "update_time") {
		err = fmt.Errorf("error")
		return
	}
	fileds, _, _ := InsertAllArgs(simple)
	if len(fileds) < 1 {
		err = fmt.Errorf("error")
		return
	}
	return
}

//UpdateSimple will update material group to database
func UpdateSimple(simple *Simple) (eff int, err error) {
	simple.UpdateTime = time.Now()
	var updateSQL string
	var where []string
	var args []interface{}

	updateSQL, args = UpdateSQL(simple, true, true, "title,image,update_time,status", "", "crud_simple")

	where, args = AppendWhere(where, args, simple.TID > 0, "tid=$%v", simple.TID)
	where, args = AppendWhere(where, args, simple.UserID > 0, "user_id=$%v", simple.UserID)

	updateSQL = JoinWhere(updateSQL, where, "and")
	fmt.Printf("update\n")
	fmt.Printf("   --->%v\n", updateSQL)
	fmt.Printf("   --->%v\n", args)

	updateSQL, args = UpdateAllSQL(simple, "crud_simple")
	fmt.Printf("update\n")
	fmt.Printf("   --->%v\n", updateSQL)
	fmt.Printf("   --->%v\n", args)

	sets, args := UpdateAllArgs(simple)
	fmt.Printf("update\n")
	fmt.Printf("   --->%v\n", sets)
	fmt.Printf("   --->%v\n", args)
	return
}

func LoadSimple(userID, simpleID int64) (simple *Simple, err error) {
	simple = &Simple{}
	var where []string
	var args []interface{}
	querySQL := QueryAllSQL(simple, "crud_simple")
	where, args = AppendWhere(where, args, true, "user_id=$%v", userID)
	where, args = AppendWhere(where, args, true, "tid=$%v", simpleID)
	querySQL = JoinWhere(querySQL, where, "and")
	fmt.Printf("query\n")
	fmt.Printf("   --->%v\n", querySQL)
	fmt.Printf("   --->%v\n", args)

	fileds := QueryAllField(simple)
	fmt.Printf("query\n")
	fmt.Printf("   --->%v\n", fileds)
	return
}

type UserIDx map[int64]string

func (u UserIDx) Scan(v interface{}) {
	s := v.(*Simple)
	u[s.TID] = fmt.Sprintf("%v-%v", s.Title, s.UserID)
}

func ScanSimple() (err error) {
	var simpleList []*Simple
	var simple *Simple
	var userIDs0, userIDs1 []int64
	var images0, images1 []*string
	var simples0 map[int64]*Simple
	var simples1 = map[int64][]*Simple{}
	var userIDm0 map[int64]int64
	var userIDm1 map[int64][]int64
	var userIDx = UserIDx{}
	err = QueryAll(
		&PoolQueryer{}, &Simple{},
		"sql", []interface{}{"arg"},
		&simpleList, &simple,
		&userIDs0, "user_id",
		&userIDs1, "user_id,all",
		&images0, "image",
		&images1, "image,all",
		&simples0, "tid",
		&simples1, "user_id",
		&userIDm0, "tid:user_id",
		&userIDm1, "user_id:tid",
		&userIDx,
		func(v *Simple) {
			// simples1[v.TID] = v
		},
	)
	if err != nil {
		return
	}
	data, _ := json.MarshalIndent(map[string]interface{}{
		"simple":      simple,
		"list":        simpleList,
		"user_ids_0":  userIDs0,
		"user_ids_1":  userIDs1,
		"images_0":    images0,
		"images_1":    images1,
		"simples_0":   simples0,
		"simples_1":   simples1,
		"user_id_m_0": userIDm0,
		"user_id_m_1": userIDm1,
		"user_id_x":   userIDx,
	}, "", "\t")
	fmt.Println("-->\n", string(data))

	return
}

func ScanError() (err error) {
	rows, _ := (&PoolQueryer{}).Query("")
	ScanAll(
		rows, &Simple{},
		func(v *Simple) {
			// simples1[v.TID] = v
		},
	)
	ScanAllArgs(&Simple{})
	//
	var images []*string
	err = QueryAll(
		&PoolQueryer{}, &Simple{},
		"sql", []interface{}{"arg"},
		&images,
	)
	if err == nil {
		err = fmt.Errorf("not error")
		return
	}
	err = QueryAll(
		&PoolQueryer{}, &Simple{},
		"sql", []interface{}{"arg"},
		&images, 1,
	)
	if err == nil {
		err = fmt.Errorf("not error")
		return
	}
	err = QueryAll(
		&PoolQueryer{}, &Simple{},
		"sql", []interface{}{"arg"},
		&images, "",
	)
	if err == nil {
		err = fmt.Errorf("not error")
		return
	}
	err = QueryAll(
		&PoolQueryer{}, &Simple{},
		"sql", []interface{}{"arg"},
		&images, "not",
	)
	if err == nil {
		err = fmt.Errorf("not error")
		return
	}

	//
	var simples map[int64]*Simple
	err = QueryAll(
		&PoolQueryer{}, &Simple{},
		"sql", []interface{}{"arg"},
		&simples,
	)
	if err == nil {
		err = fmt.Errorf("not error")
		return
	}
	err = QueryAll(
		&PoolQueryer{}, &Simple{},
		"sql", []interface{}{"arg"},
		&simples, 1,
	)
	if err == nil {
		err = fmt.Errorf("not error")
		return
	}
	err = QueryAll(
		&PoolQueryer{}, &Simple{},
		"sql", []interface{}{"arg"},
		&simples, "",
	)
	if err == nil {
		err = fmt.Errorf("not error")
		return
	}
	err = QueryAll(
		&PoolQueryer{}, &Simple{},
		"sql", []interface{}{"arg"},
		&simples, "not",
	)
	if err == nil {
		err = fmt.Errorf("not error")
		return
	}
	err = QueryAll(
		&PoolQueryer{}, &Simple{},
		"sql", []interface{}{"arg"},
		&simples, "tid:not",
	)
	if err == nil {
		err = fmt.Errorf("not error")
		return
	}

	//
	err = QueryAll(
		&PoolQueryer{}, &Simple{},
		"sql", []interface{}{"arg"},
		1,
	)
	if err == nil {
		err = fmt.Errorf("not error")
		return
	}

	//
	err = QueryAll(
		&PoolQueryer{QueryErr: fmt.Errorf("xx")}, &Simple{},
		"sql", []interface{}{"arg"},
		&simples, "tid",
	)
	if err == nil {
		err = fmt.Errorf("not error")
		return
	}
	err = QueryAll(
		&PoolQueryer{ScanErr: fmt.Errorf("xx")}, &Simple{},
		"sql", []interface{}{"arg"},
		&simples, "tid",
	)
	if err == nil {
		err = fmt.Errorf("not error")
		return
	}
	err = nil
	return
}

func TestSimple(t *testing.T) {
	//add material group
	title := "test"
	simple := &Simple{
		UserID: 100,
		Type:   "test",
		Title:  &title,
		Image:  &title,
		Status: SimpleStatusNormal,
	}
	AddSimple(simple)
	AddSimple(&Simple{
		UserID: 0,
		Type:   "test",
		Title:  &title,
		Status: SimpleStatusNormal,
	})
	UpdateSimple(simple)
	LoadSimple(simple.UserID, simple.TID)

	err := ScanSimple()
	if err != nil {
		t.Error(err)
		return
	}
	err = ScanError()
	if err != nil {
		t.Error(err)
		return
	}

}
