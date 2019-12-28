package LinkedList

import (
	"container/list"
	"reflect"
)

type linkedList struct {
	v *list.List
}


func NewLinkedList(source interface{}) *linkedList{

	l:=list.New()


	val:=reflect.ValueOf(source)

	if val.Kind()!=reflect.Slice{
		return &linkedList{}
	}else{
		for i:=0;i<val.Len();i++{
			l.PushBack(val.Index(i).Interface())
		}

		return &linkedList{
			v:l,
		}
	}



}