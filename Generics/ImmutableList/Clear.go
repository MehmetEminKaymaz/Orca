package ImmutableList

import "reflect"

func(q query) Clear() query{
	slice:=reflect.MakeSlice(reflect.SliceOf(q.v.Index(0).Type()),0,1)
	q.v=reflect.Indirect(slice)

	return q
}
