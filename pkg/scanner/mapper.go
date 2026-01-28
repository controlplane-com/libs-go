package scanner

const MapperTagName = "mapper"

var Mappers = NewMapperCollection()

type MapperCollection struct {
	registeredMappers map[string]Mapper
}

func NewMapperCollection() *MapperCollection {
	mc := &MapperCollection{registeredMappers: map[string]Mapper{}}
	mc.RegisterMapper("ConcreteMapper", ConcreteMapper{})
	mc.RegisterMapper("JsonMapper", JsonMapper{})
	return mc
}

type Mapper interface {
	Map(dest any, scannedValue any) (any, error)
}

func (mc *MapperCollection) GetMapperByName(name string) Mapper {
	return mc.registeredMappers[name]
}

func (mc *MapperCollection) RegisterMapper(name string, m Mapper) {
	mc.registeredMappers[name] = m
}
