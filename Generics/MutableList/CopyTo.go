package MutableList


import "reflect"

func(q *query) CopyTo(slice interface{}) {
	reflect.Copy(reflect.Indirect(reflect.ValueOf(slice)),q.v)

}
