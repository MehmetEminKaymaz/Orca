package MutableList


import "reflect"

func(q *query) Distinct() query{

	slice:=reflect.MakeSlice(reflect.SliceOf(q.v.Index(0).Type()),0,1)
	IsHere:=func(item interface{})(state bool){
		state =false
		for k:=0;k<slice.Len();k++{
			if slice.Index(k).Interface()==item{
				state=true
			}
		}
		return
	}
	for i:=0;i<q.v.Len();i++{
		if !IsHere(q.v.Index(i).Interface()){
			slice=reflect.Append(slice,q.v.Index(i))
		}
	}


	return query{
		v:slice,
	}
}
