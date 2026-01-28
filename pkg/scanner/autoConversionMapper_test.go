package scanner

import (
	"fmt"
	"reflect"
	"testing"
)

func TestAutoConversionMapper_Map_ConvertInt(t *testing.T) {
	a := AutoConversionMapper{}
	var s int64 = -1
	var d int = 0
	r, err := a.Map(d, s)
	if err != nil {
		t.Fail()
		return
	}
	if r != -1 {
		t.Fail()
		return
	}
}

func TestAutoConversionMapper_Map_ConvertInvalidTypes(t *testing.T) {
	a := AutoConversionMapper{}
	var s int64 = -1
	var d chan bool
	c, _ := a.Map(d, s)
	fmt.Println(c)
	if c != s {
		t.Fail()
		return
	}
}

func TestAutoConversionMapper_Map_ConvertIdentity(t *testing.T) {
	a := AutoConversionMapper{}
	var s = -1
	var d = 0
	r, err := a.Map(d, s)
	if err != nil {
		t.Fail()
		return
	}
	if r != -1 {
		t.Fail()
		return
	}
}

func TestAutoConversionMapper_Map_CreateArray(t *testing.T) {
	a := AutoConversionMapper{}
	var s = -1
	var d [5]int
	r, err := a.Map(d, s)
	if err != nil {
		t.Fail()
		return
	}
	rType := reflect.TypeOf(r)
	rValue := reflect.ValueOf(r)
	if rType.Kind() != reflect.Array {
		t.Fail()
		return
	}
	if rValue.Len() != 1 {
		t.Fail()
		return
	}
	if rValue.Index(0).Interface() != s {
		t.Fail()
		return
	}
}

func TestAutoConversionMapper_Map_CreateSlice(t *testing.T) {
	a := AutoConversionMapper{}
	var s = -1
	var d []int
	r, err := a.Map(d, s)
	if err != nil {
		t.Fail()
		return
	}
	rType := reflect.TypeOf(r)
	rValue := reflect.ValueOf(r)
	if rType.Kind() != reflect.Slice {
		t.Fail()
		return
	}
	if rValue.Len() != 1 {
		t.Fail()
		return
	}
	if rValue.Index(0).Interface() != s {
		t.Fail()
		return
	}
}
