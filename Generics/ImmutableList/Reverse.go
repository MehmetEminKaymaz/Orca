package ImmutableList

import "reflect"

func(q query) Reverse() query {

	newSlice:=reflect.MakeSlice(reflect.SliceOf(q.v.Index(0).Type()),0,1)
	for i:=q.v.Len()-1;i>=0;i--{
		newSlice=reflect.Append(newSlice,q.v.Index(i))
	}

	q.v=newSlice

	return q
}
