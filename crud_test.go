package crud

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

func JSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return err.Error()
	}
	return string(data)
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
			}, "#all"),
			ScanArgs(&Simple{
				TID:        2,
				UserID:     0,
				Type:       "test",
				Title:      &title,
				Image:      &title,
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
				Status:     SimpleStatusNormal,
			}, "#all"),
			ScanArgs(&Simple{
				TID:        2,
				UserID:     0,
				Type:       "test",
				Title:      &title,
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
				Status:     SimpleStatusNormal,
			}, "#all"),
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
	if len(dests) < 1 {
		err = fmt.Errorf("dest is empty")
		return
	}
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
	insertSQL, insertArg := InsertSQL(simple, "", "crud_simple", "returning tid")
	fmt.Printf("insert\n")
	fmt.Printf("   --->%v\n", insertSQL)
	fmt.Printf("   --->%v\n", insertArg)
	insertSQL1, insertArg1 := InsertSQL(simple, "^update_time,create_time,status", "crud_simple", "returning tid")
	fmt.Printf("insert\n")
	fmt.Printf("   --->%v\n", insertSQL1)
	fmt.Printf("   --->%v\n", insertArg1)
	if strings.Contains(insertSQL1, "update_time") {
		err = fmt.Errorf("error")
		return
	}
	fileds, _, _ := InsertArgs(simple, "")
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

	updateSQL, args = UpdateSQL(simple, "title,image,update_time,status", "crud_simple")

	if strings.Contains(updateSQL, "tid") {
		err = fmt.Errorf("error")
		return
	}

	where, args = AppendWhere(where, args, simple.TID > 0, "tid=$%v", simple.TID)
	where, args = AppendWhere(where, args, simple.UserID > 0, "user_id=$%v", simple.UserID)

	updateSQL = JoinWhere(updateSQL, where, "and")
	fmt.Printf("update\n")
	fmt.Printf("   --->%v\n", updateSQL)
	fmt.Printf("   --->%v\n", args)

	updateSQL, args = UpdateSQL(simple, "", "crud_simple")
	fmt.Printf("update\n")
	fmt.Printf("   --->%v\n", updateSQL)
	fmt.Printf("   --->%v\n", args)

	sets, args := UpdateArgs(simple, "")
	fmt.Printf("update\n")
	fmt.Printf("   --->%v\n", sets)
	fmt.Printf("   --->%v\n", args)
	return
}

func LoadSimple(userID, simpleID int64) (err error) {
	simple := &Simple{}
	var where []string
	var args []interface{}
	querySQL := QuerySQL(simple, "#nil,zero", "crud_simple")
	where, args = AppendWhere(where, args, true, "user_id=$%v", userID)
	where, args = AppendWhere(where, args, true, "tid=$%v", simpleID)
	querySQL = JoinWhere(querySQL, where, "and")
	fmt.Printf("query\n")
	fmt.Printf("   --->%v\n", querySQL)
	fmt.Printf("   --->%v\n", args)

	fileds := QueryField(simple, "#all")
	fmt.Printf("query\n")
	fmt.Printf("   --->%v\n", fileds)
	if len(fileds) < 1 {
		err = fmt.Errorf("fail")
	}
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
	err = Query(
		&PoolQueryer{}, &Simple{}, "#all",
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
	if len(simpleList) < 1 || simple == nil || len(userIDs0) < 1 || len(userIDs1) < 1 ||
		len(images0) < 1 || len(images1) < 1 || len(simples0) < 1 || len(simples1) < 1 ||
		len(userIDm0) < 1 || len(userIDm1) < 1 || len(userIDx) < 1 {
		fmt.Printf(`
			len(simpleList)=%v, simple=%v, len(userIDs0)=%v, len(userIDs1)=%v,
			len(images0)=%v, len(images1)=%v, len(simples0)=%v, len(simples1)=%v,
			len(userIDm0)=%v, len(userIDm1)=%v, len(userIDx)=%v
			`,
			len(simpleList), simple == nil, len(userIDs0), len(userIDs1),
			len(images0), len(images1), len(simples0), len(simples1),
			len(userIDm0), len(userIDm1), len(userIDx),
		)
		err = fmt.Errorf("data error")
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
	Scan(
		rows, &Simple{}, "#all",
		func(v *Simple) {
			// simples1[v.TID] = v
		},
	)
	ScanArgs(&Simple{}, "#all")
	//
	var images []*string
	err = Query(
		&PoolQueryer{}, &Simple{}, "#all",
		"sql", []interface{}{"arg"},
		&images,
	)
	if err == nil {
		err = fmt.Errorf("not error")
		return
	}
	err = Query(
		&PoolQueryer{}, &Simple{}, "#all",
		"sql", []interface{}{"arg"},
		&images, 1,
	)
	if err == nil {
		err = fmt.Errorf("not error")
		return
	}
	err = Query(
		&PoolQueryer{}, &Simple{}, "#all",
		"sql", []interface{}{"arg"},
		&images, "",
	)
	if err == nil {
		err = fmt.Errorf("not error")
		return
	}
	err = Query(
		&PoolQueryer{}, &Simple{}, "#all",
		"sql", []interface{}{"arg"},
		&images, "not",
	)
	if err == nil {
		err = fmt.Errorf("not error")
		return
	}

	//
	var simples map[int64]*Simple
	err = Query(
		&PoolQueryer{}, &Simple{}, "#all",
		"sql", []interface{}{"arg"},
		&simples,
	)
	if err == nil {
		err = fmt.Errorf("not error")
		return
	}
	err = Query(
		&PoolQueryer{}, &Simple{}, "#all",
		"sql", []interface{}{"arg"},
		&simples, 1,
	)
	if err == nil {
		err = fmt.Errorf("not error")
		return
	}
	err = Query(
		&PoolQueryer{}, &Simple{}, "#all",
		"sql", []interface{}{"arg"},
		&simples, "",
	)
	if err == nil {
		err = fmt.Errorf("not error")
		return
	}
	err = Query(
		&PoolQueryer{}, &Simple{}, "#all",
		"sql", []interface{}{"arg"},
		&simples, "not",
	)
	if err == nil {
		err = fmt.Errorf("not error")
		return
	}
	err = Query(
		&PoolQueryer{}, &Simple{}, "#all",
		"sql", []interface{}{"arg"},
		&simples, "tid:not",
	)
	if err == nil {
		err = fmt.Errorf("not error")
		return
	}

	//
	err = Query(
		&PoolQueryer{}, &Simple{}, "#all",
		"sql", []interface{}{"arg"},
		1,
	)
	if err == nil {
		err = fmt.Errorf("not error")
		return
	}

	//
	err = Query(
		&PoolQueryer{QueryErr: fmt.Errorf("xx")}, &Simple{}, "#all",
		"sql", []interface{}{"arg"},
		&simples, "tid",
	)
	if err == nil {
		err = fmt.Errorf("not error")
		return
	}
	err = Query(
		&PoolQueryer{ScanErr: fmt.Errorf("xx")}, &Simple{}, "#all",
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
	var err error
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
	_, err = UpdateSimple(simple)
	if err != nil {
		t.Error(err)
		return
	}
	err = LoadSimple(simple.UserID, simple.TID)
	if err != nil {
		t.Error(err)
		return
	}

	err = ScanSimple()
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
