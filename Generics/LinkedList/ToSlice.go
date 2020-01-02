package LinkedList

import "reflect"

func(l *linkedList) ToSlice() interface{}{

	newSlice:=make([]interface{},0)

	for e:=l.v.Front();e!=nil;e=e.Next(){
		newSlice=append(newSlice,e.Value)
	}

	return reflect.Indirect(reflect.ValueOf(newSlice))

}
