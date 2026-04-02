package scanner

import (
	"fmt"
	"github.com/lib/pq"
	"github.com/controlplane-com/libs-go/pkg/common"
	"github.com/controlplane-com/libs-go/pkg/types"
	"testing"
)

type TestTarget struct {
	IRREGULARLYNamedField       string
	Name                        string
	Description                 *string
	Map                         map[string]string
	Slice                       []string
	MultipleIndirections        **string
	Count                       int
	Sub                         SubStruct `cpln:"mapper:JsonMapper"`
	Scanner                     pq.StringArray
	PtrToScanner                *pq.StringArray
	ScannerMultipleIndirections **pq.StringArray
}

type SubStruct struct {
	Name string
}

func TestGetInvalidField(t *testing.T) {
	tt := TestTarget{}
	st := NewScanTarget(&tt)
	err := common.Try(func() error {
		st.GetFieldValue("NON_EXISTENT_FIELD")
		return nil
	})
	if err == nil {
		t.Fail()
	}
}

func TestSetInvalidField(t *testing.T) {
	tt := TestTarget{}
	st := NewScanTarget(&tt)
	err := st.SetFieldValue("NON_EXISTENT_FIELD", "")
	if err == nil {
		t.Fail()
		return
	}
}

func TestScanIntoString(t *testing.T) {
	tt := TestTarget{
		Name:        "Bob",
		Description: common.StrPtr("hello"),
		Count:       0,
		Sub: SubStruct{
			Name: "",
		},
	}

	st := NewScanTarget(&tt)
	st.AddField("Name", ConcreteMapper{})
	err := st.SetFieldValue("Name", "Kyle")
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
		return
	}
	f := st.GetFieldValue("Name").(string)
	m := st.Model.Name
	if m != f || tt.Name != f || tt.Name != m {
		t.Fail()
		return
	}
}

func TestScanIntoStringArray(t *testing.T) {
	tt := TestTarget{
		Name:        "Bob",
		Description: common.StrPtr("hello"),
		Count:       0,
		Sub: SubStruct{
			Name: "",
		},
	}

	st := NewScanTarget(&tt)
	st.AddField("Scanner", ConcreteMapper{})
	err := st.getAddedField("Scanner").Scan("{one,two,three}")
	if err != nil {
		t.Fail()
	}
	if len(st.Model.Scanner) != 3 {
		t.Fail()
	}
	if st.Model.Scanner[0] != "one" || st.Model.Scanner[1] != "two" || st.Model.Scanner[2] != "three" {
		t.Fail()
	}
}

func TestScanIntoStringArrayPtr(t *testing.T) {
	tt := TestTarget{
		Name:        "Bob",
		Description: common.StrPtr("hello"),
		Count:       0,
		Sub: SubStruct{
			Name: "",
		},
	}

	st := NewScanTarget(&tt)
	st.AddField("PtrToScanner", ConcreteMapper{})
	err := st.getAddedField("PtrToScanner").Scan("{one,two,three}")
	if err != nil {
		t.Fail()
		return
	}
	if len(*st.Model.PtrToScanner) != 3 {
		t.Fail()
		return
	}
	a := *st.Model.PtrToScanner
	if a[0] != "one" || a[1] != "two" || a[2] != "three" {
		t.Fail()
	}
}

func TestScanIntoStringArrayPtrMultipleIndirections(t *testing.T) {
	tt := TestTarget{
		Name:        "Bob",
		Description: common.StrPtr("hello"),
		Count:       0,
		Sub: SubStruct{
			Name: "",
		},
	}

	st := NewScanTarget(&tt)
	st.AddField("ScannerMultipleIndirections", ConcreteMapper{})
	err := st.SetFieldValue("ScannerMultipleIndirections", "{one,two,three}")
	if err != nil {
		t.Fail()
		return
	}
	if len(**st.Model.ScannerMultipleIndirections) != 3 {
		t.Fail()
		return
	}
	a := **st.Model.ScannerMultipleIndirections
	if a[0] != "one" || a[1] != "two" || a[2] != "three" {
		t.Fail()
	}
}

func TestScanIntoStringPtr(t *testing.T) {
	tt := TestTarget{
		Name:        "Bob",
		Description: common.StrPtr("hello"),
		Count:       0,
		Sub: SubStruct{
			Name: "",
		},
	}

	st := NewScanTarget(&tt)
	st.AddField("Description", ConcreteMapper{})
	err := st.SetFieldValue("Description", "new description")
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
		return
	}
	ps := st.GetFieldValue("Description").(*string)
	pm := st.Model.Description
	if *pm != "new description" || *ps != "new description" || pm != ps {
		t.Fail()
	}

	s := "ptr description"
	err = st.SetFieldValue("Description", &s)
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
		return
	}
	ps = st.GetFieldValue("Description").(*string)
	pm = st.Model.Description
	if *pm != "ptr description" || *ps != "ptr description" || pm != ps {
		t.Fail()
	}

	strPtr := &s
	err = st.SetFieldValue("Description", &strPtr)
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
		return
	}
	ps = st.GetFieldValue("Description").(*string)
	pm = st.Model.Description
	if *pm != "ptr description" || *ps != "ptr description" || pm != ps {
		t.Fail()
	}

	strPtr = nil
	err = st.SetFieldValue("Description", &strPtr)
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
		return
	}
	ps = st.GetFieldValue("Description").(*string)
	pm = st.Model.Description
	if *pm != "" || *ps != "" {
		t.Fail()
	}

	var strPtrPtr **string
	err = st.SetFieldValue("Description", strPtrPtr)
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
		return
	}
	ps = st.GetFieldValue("Description").(*string)
	pm = st.Model.Description
	if *pm != "" || *ps != "" {
		t.Fail()
	}
}

func TestScanNilIntoPtr(t *testing.T) {
	tt := TestTarget{
		Name:        "Bob",
		Description: common.StrPtr("hello"),
		Count:       0,
		Sub: SubStruct{
			Name: "",
		},
	}

	st := NewScanTarget(&tt)
	st.AddField("Description", nil)
	var s *string
	err := st.SetFieldValue("Description", s)
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
		return
	}
	ps := st.GetFieldValue("Description").(*string)
	pm := st.Model.Description
	if pm != nil || ps != nil {
		t.Fail()
	}
}

func TestScanIntoNilMultiLevelPtr(t *testing.T) {
	tt := TestTarget{}

	st := NewScanTarget(&tt)
	st.AddField("MultipleIndirections", nil)
	var s *string
	err := st.SetFieldValue("MultipleIndirections", s)
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
		return
	}
	ps := st.GetFieldValue("MultipleIndirections").(**string)
	pm := st.Model.Description
	if pm != nil || *ps != nil {
		t.Fail()
	}
}

func TestScanIncorrectIndirectionMismatch(t *testing.T) {
	tt := TestTarget{
		Name:        "Bob",
		Description: common.StrPtr("hello"),
		Count:       0,
		Sub: SubStruct{
			Name: "",
		},
	}

	st := NewScanTarget(&tt)
	st.AddField("Description", nil)
	var s **string
	err := st.SetFieldValue("Description", s)
	if err == nil {
		t.Fail()
		return
	}
}

func TestScanCorrectIndirectionMismatch(t *testing.T) {
	tt := TestTarget{
		Name:                 "Bob",
		MultipleIndirections: types.CopyAndIndirect(2, "hello").(**string),
		Count:                0,
		Sub: SubStruct{
			Name: "",
		},
	}

	st := NewScanTarget(&tt)
	st.AddField("MultipleIndirections", nil)
	err := st.SetFieldValue("MultipleIndirections", "test")
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
		return
	}
	f := st.GetFieldValue("MultipleIndirections").(**string)
	m := st.Model.MultipleIndirections
	if **m != "test" || **f != "test" {
		t.Fail()
	}
}

func TestScanCorrectIndirectionMismatchPartialNil(t *testing.T) {
	var sp *string
	tt := TestTarget{
		Name:                 "Bob",
		MultipleIndirections: &sp,
		Count:                0,
		Sub: SubStruct{
			Name: "",
		},
	}

	st := NewScanTarget(&tt)
	st.AddField("MultipleIndirections", nil)
	err := st.SetFieldValue("MultipleIndirections", "test")
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
		return
	}
	f := st.GetFieldValue("MultipleIndirections").(**string)
	m := st.Model.MultipleIndirections
	if **m != "test" || **f != "test" {
		t.Fail()
	}
}

func TestScanNilIntoString(t *testing.T) {
	tt := TestTarget{
		Name:  "Bob",
		Count: 0,
		Sub: SubStruct{
			Name: "",
		},
	}

	st := NewScanTarget(&tt)
	st.AddField("Name", ConcreteMapper{})
	err := st.SetFieldValue("Name", nil)
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
		return
	}
	f := st.GetFieldValue("Name").(string)
	m := st.Model.Name
	if m != "" || f != "" {
		t.Fail()
	}
}

func TestMapperOverride(t *testing.T) {
	tt := TestTarget{
		Sub: SubStruct{
			Name: "",
		},
	}

	st := NewScanTarget(&tt)
	st.AddField("Sub", nil)
	err := st.SetFieldValue("Sub", `{"Name":"Kyle"}`)
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
		return
	}
	f := st.GetFieldValue("Sub").(SubStruct)
	m := st.Model.Sub
	if m.Name != "Kyle" || f.Name != "Kyle" {
		t.Fail()
	}
}

func TestMapperErrorOverride(t *testing.T) {
	tt := TestTarget{
		Sub: SubStruct{
			Name: "",
		},
	}

	st := NewScanTarget(&tt)
	st.AddField("Sub", nil)
	err := st.SetFieldValue("Sub", `{badSyntax":"hi"}`)
	if err == nil {
		fmt.Println(err.Error())
		t.Fail()
		return
	}
}

func TestAddInvalidScanTargetField(t *testing.T) {
	tt := TestTarget{}
	st := NewScanTarget(&tt)
	st.AddField("INVALID_FIELD", nil)
	if st.GetFieldValue("INVALID_FIELD") != "" {
		t.Fail()
	}
	_ = st.SetFieldValue("INVALID_FIELD", "hello")
	if st.GetFieldValue("INVALID_FIELD") != "" {
		t.Fail()
	}
}

func TestAddIrregularlyNamedField(t *testing.T) {
	tt := TestTarget{}
	st := NewScanTarget(&tt)
	st.AddField("IrregularlyNamedField", nil)
	if st.GetFieldValue("IrregularlyNamedField") != "" {
		t.Fail()
	}
	_ = st.SetFieldValue("IrregularlyNamedField", "hello")
	if st.GetFieldValue("IrregularlyNamedField") != "hello" {
		t.Fail()
	}
	if st.GetFieldValue("IrReGuLarlYNamedField") != "hello" {
		t.Fail()
	}
	if st.GetFieldValue("IrReGuLarlY_Named-Field") != "hello" {
		t.Fail()
	}
	if st.Model.IRREGULARLYNamedField != "hello" {
		t.Fail()
	}
}

func TestScanTarget_HasField(t *testing.T) {
	tt := TestTarget{}
	st := NewScanTarget(&tt)
	st.AddField("Name", nil)

	if st.FieldHasBeenAdded("NOT_ADDED") {
		t.Fail()
	}
	if !st.FieldHasBeenAdded("Name") {
		t.Fail()
	}
}

func TestAddInvalidScanTarget(t *testing.T) {
	tt := TestTarget{}
	err := common.Try(func() error {
		NewScanTarget(tt)
		return nil
	})
	if err == nil {
		t.Fail()
	}
}

func TestAssignNilLiteral(t *testing.T) {
	tt := TestTarget{
		Sub: SubStruct{
			Name: "Kyle",
		},
	}

	st := NewScanTarget(&tt)

	//Name cannot accept nil because it's a string
	st.AddField("Name", nil)
	err := st.SetFieldValue("Name", nil)
	if err == nil {
		t.Fail()
		return
	}

	//Description can accept nil because it's a *string
	st.AddField("Description", nil)
	err = st.SetFieldValue("Description", nil)
	if err != nil {
		t.Fail()
		return
	}

	//Map can accept nil because it's a map
	st.AddField("Map", nil)
	err = st.SetFieldValue("Map", nil)
	if err != nil {
		t.Fail()
		return
	}

	//Slice can accept nil because it's a slice
	st.AddField("Slice", nil)
	err = st.SetFieldValue("Slice", nil)
	if err != nil {
		t.Fail()
		return
	}
}

func TestScanTargetCopy(t *testing.T) {
	tt := TestTarget{
		Name:  "Bob",
		Count: 0,
		Sub: SubStruct{
			Name: "",
		},
	}

	st := CopyIntoNewScanTarget(&tt)
	st.AddField("Name", ConcreteMapper{})
	err := st.SetFieldValue("Name", "Kyle")
	if err != nil {
		t.Fail()
		return
	}
	f := st.GetFieldValue("Name")
	m := st.Model.Name
	if m != f || m == tt.Name || f == tt.Name {
		t.Fail()
	}
}
