package ImmutableList

import "reflect"

func(q query) Foreach(do func(x interface{}) (y interface{})) query {
	for i:=0;i<q.v.Len();i++{
		q.v.Index(i).Set(reflect.ValueOf(do(q.v.Index(i).Interface())))
	}
	return q
}
