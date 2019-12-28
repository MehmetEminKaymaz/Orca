package LinkedList

import "reflect"

func(l *linkedList) Exist(slice interface{}) (state bool){

	slicev:=reflect.ValueOf(slice)
	state=true

	for i:=0;i<slicev.Len();i++{

		for e:=l.v.Front();e!=nil;e=e.Next(){

			if e.Value!=slicev.Index(i).Interface(){
				state = false
				return
			}

		}

	}

	return

}
