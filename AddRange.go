package Orca

import "reflect"

func(q *Query) AddRange(slice interface{}){
	slicev:=reflect.ValueOf(slice)
	for i:=0;i<slicev.Len();i++{
		q.v=reflect.Append(q.v,slicev.Index(i))
	}

}
