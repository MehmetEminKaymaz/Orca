package ImmutableList

import "reflect"

type query struct {

	v reflect.Value

}


func From(source interface{}) query{

	val:=reflect.ValueOf(source)
	if val.Kind()!=reflect.Slice{
		return query{}
	}
	return query{
		v:val,
	}

}

func(q query) Len() int{
	return q.v.Len()
}

func(q query) ToSlice() interface{}{
	return reflect.Indirect(q.v)
}
