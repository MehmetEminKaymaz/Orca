package LazyList

import (
	"reflect"
	"errors"
)

type task struct {
	KindOfTask string
	Do         func() interface{}
}



type empty struct {
}

type collection struct {
	Source reflect.Value
	Tasks  []task
	Items  []interface{}
}

func NewLazyList(source interface{}) (collection, error) {

	val := reflect.ValueOf(source)
	switch val.Kind() {

	case reflect.Slice, reflect.Array:

		myMap := make([]task, 0, 1)
		myItems := make([]interface{}, 0, 1)

		return collection{
			Source: reflect.ValueOf(source),
			Tasks:  myMap,
			Items:  myItems,
		}, nil

	default:

		return collection{}, errors.New("It is not slice or array")

	}

}