package MutableList


import "reflect"

func(q query) Skip( skp int) query{

	newSlice:=reflect.MakeSlice(reflect.SliceOf(q.v.Index(0).Type()),0,1)
	for i:=skp;i<q.v.Len();i++{
		newSlice=reflect.Append(newSlice,q.v.Index(i))
	}
	return query{
		v:newSlice,
	}

}
