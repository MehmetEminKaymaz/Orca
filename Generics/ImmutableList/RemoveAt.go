package ImmutableList

import "reflect"

func(q query) RemoveAt(index int) query {


	newSlice:=reflect.MakeSlice(reflect.SliceOf(q.v.Index(0).Type()),0,1)

	for i:=0;i<q.v.Len();i++{
		if index!=i{
			newSlice=reflect.Append(newSlice,q.v.Index(i))
		}

	}
	q.v=newSlice

   return q

}