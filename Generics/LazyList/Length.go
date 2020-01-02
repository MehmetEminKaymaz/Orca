package LazyList


import "reflect"

func (c collection) Length() int { //len(collection)
	return reflect.Indirect(c.Source).Len()
}

func (c collection) Value() interface{} { //source to slice

	return c.Source
}