package MutableList

import "reflect"

func(q *query) Add(x interface{}) {
	q.v=reflect.Append(q.v,reflect.ValueOf(x))
}

