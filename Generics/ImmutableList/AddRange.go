package ImmutableList

import "reflect"

func(q query) AddRange(slice interface{}) query {
	slicev:=reflect.ValueOf(slice)
	for i:=0;i<slicev.Len();i++{
		q.v=reflect.Append(q.v,slicev.Index(i))
	}

	return q

}