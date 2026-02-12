package lib

import (
	"reflect"
	"time"

	"github.com/my-easy-vault-2026/api-server/utils"
)

type DatabaseTimezoneLib struct {
	logger Logger
	env    *Env
}

func NewDatabaseTimezoneLib(logger Logger, env *Env) *DatabaseTimezoneLib {
	return &DatabaseTimezoneLib{
		logger: logger,
		env:    env,
	}
}

// 插入,更新,查詢時間時請使用這個函數
func (dtl *DatabaseTimezoneLib) DBQueryTime(t time.Time) time.Time {
	_, offset := time.Now().Zone()
	offsetHour := offset / 60 / 60

	dbOffsetHour := dtl.env.DBTimezoneOffset

	return t.Add(time.Hour * time.Duration((dbOffsetHour - offsetHour)))
}

// DB查出的時間，轉換至伺服器時間
func (dtl *DatabaseTimezoneLib) TimeFromDB(t time.Time) time.Time {
	_, offset := time.Now().Zone()
	offsetHour := offset / 60 / 60

	dbOffsetHour := dtl.env.DBTimezoneOffset
	return t.Add(time.Hour * time.Duration((offsetHour - dbOffsetHour)))
}

func (dtl *DatabaseTimezoneLib) ObjectFromDB(o interface{}) interface{} {
	if reflect.TypeOf(o).Kind() != reflect.Pointer || reflect.ValueOf(o).IsNil() {
		return o
	}
	structType := reflect.ValueOf(o).Elem().Type()
	structValue := reflect.ValueOf(o).Elem()
	structPtrValue := reflect.ValueOf(o)
	for i := 0; i < structType.NumField(); i++ {
		if structValue.Field(i).IsZero() {
			continue
		}
		kind := structValue.Field(i).Kind()
		switch kind {
		case reflect.Pointer:
			ptr := structValue.Field(i).Interface()
			switch reflect.TypeOf(ptr) {
			case reflect.TypeOf((*time.Time)(nil)):
				t := ptr.(*time.Time)
				if t.IsZero() {
					continue
				}
				structPtrValue.Elem().Field(i).Set(reflect.ValueOf(utils.Ptr(dtl.TimeFromDB(*t))))
			default:
				continue
			}
		case reflect.Struct:
			ptr := structPtrValue.Elem().Field(i).Addr().Interface()
			switch reflect.TypeOf(ptr) {
			case reflect.TypeOf((*time.Time)(nil)):
				t := ptr.(*time.Time)
				if t.IsZero() {
					continue
				}
				structPtrValue.Elem().Field(i).Set(reflect.ValueOf(dtl.TimeFromDB(*t)))
			default:
				continue
			}
		default:
			continue
		}
	}
	return o
}
