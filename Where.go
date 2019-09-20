package Orca

import "reflect"

func (q Query) Where(ok func(x interface{}) bool) Query{

	newSlice:=reflect.MakeSlice(reflect.SliceOf(q.v.Index(0).Type()),0,1)

	for i:=0;i<q.v.Len();i++{
		if ok(q.v.Index(i).Interface()){
			newSlice=reflect.Append(newSlice,q.v.Index(i))
		}
	}

	return Query{
		v:newSlice,
	}
}
