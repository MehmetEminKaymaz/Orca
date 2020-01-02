package MutableList

import "reflect"

type query struct {

   v reflect.Value

}



func NewMutableList(source interface{}) query{
	val:=reflect.ValueOf(source)
	if val.Kind()!=reflect.Slice{
		return query{}
	}
	return query{
		v:val,
	}
}

func(q *query) Len() int{
	return q.v.Len()
}


func(q query) ToSlice() (slice interface{}){
	return reflect.Indirect(q.v)
}
