package ImmutableList

import "reflect"

func(q query) Add(x interface{}) query {
	q.v=reflect.Append(q.v,reflect.ValueOf(x))
	return  q
}
