package Orca

import "reflect"

func(q Query) Take(tk int) Query{
	newSlice:=reflect.MakeSlice(reflect.SliceOf(q.v.Index(0).Type()),0,1)
	if tk>q.v.Len(){
		tk=q.v.Len()
	}
	for i:=0;i<tk;i++{
		newSlice=reflect.Append(newSlice,q.v.Index(i))
	}

	return Query{
		v:newSlice,
	}
}
