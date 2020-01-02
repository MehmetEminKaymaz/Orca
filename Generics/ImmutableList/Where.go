package ImmutableList

import "reflect"

func (q query) Where(ok func(x interface{}) bool) query{

	newSlice:=reflect.MakeSlice(reflect.SliceOf(q.v.Index(0).Type()),0,1)

	for i:=0;i<q.v.Len();i++{
		if ok(q.v.Index(i).Interface()){
			newSlice=reflect.Append(newSlice,q.v.Index(i))
		}
	}

	return query{
		v:newSlice,
	}
}
