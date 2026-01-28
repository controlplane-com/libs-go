package scanner

import (
	"errors"
	"fmt"
	"github.com/controlplane-com/libs-go/pkg/metadata"
	"github.com/controlplane-com/libs-go/pkg/types"
	"github.com/iancoleman/strcase"
	"reflect"
	"strings"
)

type Scanner interface {
	Scan(src any) error
}

var ErrNoSuchField = errors.New("no such field")

func CopyIntoNewScanTarget[T any](model T) *ScanTarget[T] {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	v := reflect.New(t).Elem()
	return &ScanTarget[T]{
		Model:           v.Addr().Interface().(T),
		mc:              Mappers,
		fields:          []*ScanTargetField{},
		addedFieldNames: map[string]int{},
		modelFieldNames: collectScanTargetModelFields(v.Type()),
		v:               v,
	}
}

func collectScanTargetModelFields(rt reflect.Type) map[string]int {
	modelFields := map[string]int{}
	for i := 0; i < rt.NumField(); i++ {
		modelFields[strings.ToUpper(rt.Field(i).Name)] = i
	}
	return modelFields
}

func NewScanTarget[T any](model T) *ScanTarget[T] {
	if reflect.ValueOf(model).Kind() != reflect.Pointer {
		panic("The given model must be a pointer")
	}
	rv := reflect.ValueOf(model).Elem()
	if rv.Kind() != reflect.Struct {
		panic("The given model must be a pointer to a struct")
	}
	return &ScanTarget[T]{
		Model:           model,
		mc:              Mappers,
		fields:          []*ScanTargetField{},
		addedFieldNames: map[string]int{},
		modelFieldNames: collectScanTargetModelFields(rv.Type()),
		v:               rv,
	}
}

type ScanTarget[T any] struct {
	Model           T
	mc              *MapperCollection
	fields          []*ScanTargetField
	modelFieldNames map[string]int
	addedFieldNames map[string]int
	v               reflect.Value
}

func (t *ScanTarget[T]) GetReflectedModel() reflect.Value {
	return t.v
}

func (t *ScanTarget[T]) FieldHasBeenAdded(name string) bool {
	return t.getModelField(name) != nil
}

func (t *ScanTarget[T]) getModelField(name string) *reflect.StructField {
	normalizedName := strings.ToUpper(strcase.ToCamel(name))
	if i, ok := t.modelFieldNames[normalizedName]; ok {
		f := t.v.Type().FieldByIndex([]int{i})
		return &f
	}
	return nil
}

func (t *ScanTarget[T]) AddAllFields(defaultMapper Mapper) {
	modelValue := t.v
	for i := 0; i < modelValue.NumField(); i++ {
		t.AddField(t.v.Type().Field(i).Name, defaultMapper)
	}
}

func (t *ScanTarget[T]) AddField(name string, defaultMapper Mapper) {
	mf := t.getModelField(name)
	var v reflect.Value
	if mf != nil {
		v = t.v.FieldByName(mf.Name)
		name = mf.Name
	}
	m := t.overrideMapper(name, defaultMapper)
	t.fields = append(t.fields, newScanTargetField(m, v, name))
	t.addedFieldNames[name] = len(t.fields) - 1
}

func (t *ScanTarget[T]) getAddedField(name string) *ScanTargetField {
	mf := t.getModelField(name)
	if mf != nil {
		name = mf.Name
	}
	if i, ok := t.addedFieldNames[name]; ok {
		return t.fields[i]
	}
	return nil
}

func (t *ScanTarget[T]) SetFieldValue(fieldName string, fieldValue any) error {
	f := t.getAddedField(fieldName)
	if f == nil {
		return ErrNoSuchField
	}
	return f.Scan(fieldValue)
}

func (t *ScanTarget[T]) GetFieldValue(fieldName string) any {
	f := t.getAddedField(fieldName)
	if f == nil {
		panic("no such field")
	}
	return f.Read()
}

func (t *ScanTarget[T]) overrideMapper(name string, defaultMapper Mapper) Mapper {
	if tag := metadata.ParseCplnTag(t.v.Type(), name); tag != nil {
		if mapperFromTag, ok := tag[MapperTagName]; ok {
			if newMapper := t.mc.GetMapperByName(mapperFromTag); newMapper != nil {
				return newMapper
			}
		}
	}
	return defaultMapper
}

type ScanTargetField struct {
	dest          reflect.Value
	concreteType  reflect.Type
	name          string
	indirectLevel int
	Mapper
}

func (f *ScanTargetField) tryDelegateScanToDest(src any) error {
	concreteDest := f.dest
	if f.indirectLevel > 0 {
		concreteDest, _ = types.EnsureConcreteValue(f.dest)
	}
	scanner, _ := types.AsInterface[Scanner](concreteDest)
	return scanner.Scan(src)
}

func (f *ScanTargetField) Scan(src any) error {
	if !f.dest.CanSet() {
		return nil
	}
	mappedValue := src
	var err error
	if f.Mapper != nil {
		mappedValue, err = f.Map(f.dest.Interface(), src)
		if err != nil {
			return errors.New(fmt.Sprintf("Error scanning field %s. Details: %v", f.name, err))
		}
	}
	var ct reflect.Type
	var t = reflect.TypeOf(mappedValue)
	if t == nil {
		t = f.dest.Type()
		mappedValue, err = types.GetNilValue(t)
		if err != nil {
			return err
		}
	}

	if types.TypeImplementsInterface[Scanner](f.concreteType) {
		return f.tryDelegateScanToDest(src)
	}

	var l int
	if ct, l = types.FollowPointersToConcreteType(t); !ct.AssignableTo(f.concreteType) || l > f.indirectLevel {
		return errors.New(fmt.Sprintf("value of type %s cannot be assigned to a field of type %s", types.GetTypeName(t), types.GetTypeName(f.dest.Type())))
	}
	correctedDest := f.correctIndirectionMismatch(t)
	correctedDest.Set(reflect.ValueOf(mappedValue))
	return nil
}

/*
 * This assumes:
 * 1. That f.dest and srcType eventually point to the same concrete type
 * 2. That f.dest is at a level of indirection >= that of srcType
 */
func (f *ScanTargetField) correctIndirectionMismatch(srcType reflect.Type) *reflect.Value {
	r := f.dest
	//Walk down levels of indirection until we find a nil value, or a non-pointer
	for r.Type() != srcType {
		if r.IsNil() {
			i := types.CountIndirectionsBetweenTypesUnsafely(r.Type(), srcType)
			d := types.CopyAndIndirect(i, reflect.New(srcType).Elem().Interface())
			r.Set(reflect.ValueOf(d))
		}
		r = r.Elem()
	}
	return &r
}

func (f *ScanTargetField) Read() any {
	return f.dest.Interface()
}

func newScanTargetField(m Mapper, fd reflect.Value, n string) *ScanTargetField {
	//Handle invalid fields gracefully
	if !fd.IsValid() {
		fd = reflect.ValueOf("")
	}
	t, l := types.FollowPointersToConcreteType(fd.Type())
	return &ScanTargetField{
		Mapper:        m,
		dest:          fd,
		name:          n,
		concreteType:  t,
		indirectLevel: l,
	}
}
