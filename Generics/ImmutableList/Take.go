package ImmutableList

import "reflect"

func(q query) Take(tk int) query{
	newSlice:=reflect.MakeSlice(reflect.SliceOf(q.v.Index(0).Type()),0,1)
	if tk>q.v.Len(){
		tk=q.v.Len()
	}
	for i:=0;i<tk;i++{
		newSlice=reflect.Append(newSlice,q.v.Index(i))
	}

	return query{
		v:newSlice,
	}
}