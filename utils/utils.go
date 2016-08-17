package utils

import (
	"crypto/sha1"
	"io"
	"fmt"
	"crypto/md5"
	"reflect"
)

//获取sha1
func GetSHA1(data string) string {
	t := sha1.New()
	io.WriteString(t, data)
	return fmt.Sprintf("%x", t.Sum(nil))
}

//获取md5
func GetMD5(data string) string {
	t := md5.New()
	io.WriteString(t, data)
	return fmt.Sprintf("%x", t.Sum(nil))
}

//数据结构对拷
//支持 struct,slice
func Copy(src interface{}, des interface{}) (err error) {
	var (
		isSlice bool
		fromType reflect.Type
		isFromPtr bool
		toType reflect.Type
		isToPtr bool
		amount int
	)

	from := reflect.Indirect(reflect.ValueOf(src))
	to := reflect.Indirect(reflect.ValueOf(des))

	if to.Kind() == reflect.Slice {
		isSlice = true
		if from.Kind() == reflect.Slice {
			fromType = from.Type().Elem()
			if fromType.Kind() == reflect.Ptr {
				fromType = fromType.Elem()
				isFromPtr = true
			}
			amount = from.Len()
		} else {
			fromType = from.Type()
			amount = 1
		}

		toType = to.Type().Elem()
		if toType.Kind() == reflect.Ptr {
			toType = toType.Elem()
			isToPtr = true
		}
	} else {
		fromType = from.Type()
		toType = to.Type()
		amount = 1
	}

	for e := 0; e < amount; e++ {
		var dest, source reflect.Value
		if isSlice {
			if from.Kind() == reflect.Slice {
				source = from.Index(e)
				if isFromPtr {
					source = source.Elem()
				}
			} else {
				source = from
			}
		} else {
			source = from
		}

		if isSlice {
			dest = reflect.New(toType).Elem()
		} else {
			dest = to
		}

		for _, field := range deepFields(fromType) {
			if !field.Anonymous {
				name := field.Name
				fromField := source.FieldByName(name)
				toField := dest.FieldByName(name)
				toMethod := dest.Addr().MethodByName(name)

				canCopy := fromField.IsValid() && toField.IsValid() &&
				toField.CanSet() && fromField.Type().AssignableTo(toField.Type())

				if canCopy {
					toField.Set(fromField)
				}

				canCopy = fromField.IsValid() && toMethod.IsValid() &&
				fromField.Type().AssignableTo(toMethod.Type().In(0))

				if canCopy {
					toMethod.Call([]reflect.Value{fromField})
				}
			}
		}

		for i := 0; i < toType.NumField(); i++ {
			field := toType.Field(i)
			if !field.Anonymous {
				name := field.Name
				fromMethod := source.Addr().MethodByName(name)
				toField := dest.FieldByName(name)

				if fromMethod.IsValid() && toField.IsValid() && toField.CanSet() {
					values := fromMethod.Call([]reflect.Value{})
					if len(values) >= 1 {
						toField.Set(values[0])
					}
				}
			}
		}

		if isSlice {
			if isToPtr {
				to.Set(reflect.Append(to, dest.Addr()))
			} else {
				to.Set(reflect.Append(to, dest))
			}
		}
	}
	return
}

func deepFields(fType reflect.Type) []reflect.StructField {
	var fields []reflect.StructField

	for i := 0; i < fType.NumField(); i++ {
		v := fType.Field(i)
		if v.Anonymous && v.Type.Kind() == reflect.Struct {
			fields = append(fields, deepFields(v.Type)...)
		} else {
			fields = append(fields, v)
		}
	}

	return fields
}
