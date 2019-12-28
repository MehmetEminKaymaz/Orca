package LinkedList

import "container/list"

func(l *linkedList) Distinct() *linkedList {

   newList :=list.New()

   wrapper:=linkedList{
   	v:newList,
   }

   for e:=l.v.Front();e!=nil;e=e.Next(){

   	if !wrapper.Contains(e.Value){
   		wrapper.Add(e.Value)
	}

   }

   l=&wrapper

   return  l

}
