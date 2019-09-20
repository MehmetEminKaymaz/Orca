package Orca


import "reflect"

func(q Query) Skip( skp int) Query{

	newSlice:=reflect.MakeSlice(reflect.SliceOf(q.v.Index(0).Type()),0,1)
	for i:=skp;i<q.v.Len();i++{
		newSlice=reflect.Append(newSlice,q.v.Index(i))
	}
	return Query{
		v:newSlice,
	}

}
