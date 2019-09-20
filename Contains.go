package Orca


import "reflect"

func(q *Query) Contains(item interface{}) (state bool){
	state=false
	for i:=0;i<q.v.Len();i++{
		if reflect.ValueOf(item).Interface()==q.v.Index(i).Interface(){
			state=true
			return
		}
	}
	return false
}
