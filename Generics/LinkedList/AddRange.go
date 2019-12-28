package LinkedList

import "reflect"

func(l *linkedList) AddRange(slice interface{}) *linkedList {

	slicev:=reflect.ValueOf(slice)

	for i:=0;i<slicev.Len();i++{
		l.v.PushBack(slicev.Index(i).Interface())
	}

	return  l


}