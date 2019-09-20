package Orca

import "reflect"

func(q *Query) Foreach(do func(x interface{}) (y interface{})){
	for i:=0;i<q.v.Len();i++{
		q.v.Index(i).Set(reflect.ValueOf(do(q.v.Index(i).Interface())))
	}
}
