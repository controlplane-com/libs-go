package types

import (
	"fmt"
	"github.com/controlplane-com/libs-go/pkg/common"
	"reflect"
	"testing"
)

func TestGetTypeName(t *testing.T) {
	var s ****string
	tn := GetTypeName(reflect.TypeOf(s))
	if tn != "****string" {
		fmt.Println(tn)
		t.Fail()
	}
}

func TestGetNilValue(t *testing.T) {
	type testI interface {
		Pub() error
	}
	type testS struct {
	}
	var (
		sp    *string
		c     chan string
		s     []string
		i     testI
		n     int
		stru  testS
		pStru *testS
	)
	types := []struct {
		t         any
		expectErr bool
	}{
		{sp, false},
		{c, false},
		{s, false},
		{i, true},
		{n, true},
		{stru, true},
		{pStru, false},
	}
	for i, test := range types {
		err := common.Try(func() error {
			_, getNilErr := GetNilValue(reflect.TypeOf(test.t))
			return getNilErr
		})
		if err != nil != test.expectErr {
			t.Fail()
			fmt.Println(fmt.Sprintf("Test %d failed", i))
		}
	}
}

func TestMultiIndirects(t *testing.T) {
	i := CopyAndIndirect(2, "string")
	if _, ok := i.(**string); !ok {
		t.Fail()
	}
}

func TestSetAfterIndirect(t *testing.T) {
	s := ""
	sp := &s
	i, _ := Indirect(2, &s)
	ptr := i.(**string)
	*sp = "hello"
	if **ptr != "hello" {
		t.Fail()
		return
	}
}

func TestSetIndirectInvalidValue(t *testing.T) {
	_, err := Indirect(2, "")
	if err == nil {
		t.Fail()
		return
	}
}

func TestEnsureConcreteValue(t *testing.T) {
	var ipp **int
	cv, err := EnsureConcreteValue(reflect.ValueOf(ipp))
	if err == nil {
		t.Fail()
		return
	}

	cv, err = EnsureConcreteValue(reflect.ValueOf(&ipp).Elem())
	if err != nil {
		t.Fail()
		return
	}
	if cv.Interface() != 0 || **ipp != 0 {
		t.Fail()
		return
	}
	cv.Set(reflect.ValueOf(1))
	if **ipp != 1 {
		t.Fail()
		return
	}
}

type SomeInterface interface {
	Method() error
}

type someInterface struct {
}

func (s someInterface) Method() error {
	return nil
}

type someInterfacePtrReceiver struct {
}

func (s *someInterfacePtrReceiver) Method() error {
	return nil
}

type nonInterface struct {
}

func TestAsInterface(t *testing.T) {
	r := reflect.ValueOf(someInterface{})
	_, err := AsInterface[SomeInterface](r)
	if err != nil {
		t.Fail()
		return
	}

	var p *someInterface
	_, err = AsInterface[SomeInterface](reflect.ValueOf(&p).Elem())
	if err != nil {
		t.Fail()
		return
	}

	var nilP *someInterface
	_, err = AsInterface[SomeInterface](reflect.ValueOf(nilP))
	if err == nil {
		t.Fail()
		return
	}

	ptrR := reflect.ValueOf(&someInterfacePtrReceiver{}).Elem()
	_, err = AsInterface[SomeInterface](ptrR)
	if err != nil {
		t.Fail()
		return
	}

	_, err = AsInterface[SomeInterface](reflect.ValueOf(someInterfacePtrReceiver{}))
	if err == nil {
		t.Fail()
		return
	}

	var ptrP *someInterfacePtrReceiver
	_, err = AsInterface[SomeInterface](reflect.ValueOf(&ptrP).Elem())
	if err != nil {
		t.Fail()
		return
	}

	var nonScanner nonInterface
	_, err = AsInterface[SomeInterface](reflect.ValueOf(&nonScanner).Elem())
	if err == nil {
		t.Fail()
		return
	}
}

func TestCountIndirectionsBetweenTypes(t *testing.T) {
	var (
		sppp  ***string
		spp   **string
		sp    *string
		s     string
		ippp  ***int
		ipp   **int
		ip    *int
		i     int
		spppt = reflect.TypeOf(sppp)
		sppt  = reflect.TypeOf(spp)
		spt   = reflect.TypeOf(sp)
		st    = reflect.TypeOf(s)
		ipppt = reflect.TypeOf(ippp)
		ippt  = reflect.TypeOf(ipp)
		ipt   = reflect.TypeOf(ip)
		it    = reflect.TypeOf(i)
	)
	var tests = []struct {
		index                int
		t1                   reflect.Type
		t2                   reflect.Type
		expcetedIndirections int
		expectErr            bool
	}{
		{0, spppt, sppt, 1, false},
		{1, spppt, spt, 2, false},
		{2, spppt, st, 3, false},
		{3, sppt, spt, 1, false},
		{4, sppt, st, 2, false},
		{5, spt, st, 1, false},
		{6, st, spt, 1, true},
		{7, st, sppt, 1, true},
		{8, spt, sppt, 1, true},
		{9, ipppt, ippt, 1, false},
		{10, ipppt, ipt, 2, false},
		{11, ipppt, it, 3, false},
		{12, ippt, ipt, 1, false},
		{13, ippt, it, 2, false},
		{14, ipt, it, 1, false},
		{15, it, ipt, 1, true},
		{16, it, ippt, 1, true},
		{17, ipt, ippt, 1, true},
		{18, spppt, ipppt, 0, true},
		{19, sppt, ippt, 0, true},
		{20, spt, ipt, 0, true},
		{21, st, it, 0, true},
	}
	for _, test := range tests {
		indirections, err := CountIndirectionsBetweenTypes(test.t1, test.t2)
		if !test.expectErr && indirections != test.expcetedIndirections {
			fmt.Println(fmt.Sprintf("Test %d failed", test.index))
			t.Fail()
		}
		if err != nil != test.expectErr {
			fmt.Println(fmt.Sprintf("Test %d failed", test.index))
			t.Fail()
		}

		err = common.Try(func() error {
			indirections = CountIndirectionsBetweenTypesUnsafely(test.t1, test.t2)
			return nil
		})
		if !test.expectErr && indirections != test.expcetedIndirections {
			fmt.Println(fmt.Sprintf("Test %d failed", test.index))
			t.Fail()
		}
		if err != nil != test.expectErr {
			fmt.Println(fmt.Sprintf("Test %d failed", test.index))
			t.Fail()
		}
	}
}
