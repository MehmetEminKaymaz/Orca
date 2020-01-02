package LazyList

import "reflect"

func (c collection) Distinct() collection { //get distinct items from source
	t := task{
		KindOfTask: "Void2",
		Do: func() interface{} {
			return func(col *collection) {

				slice := reflect.MakeSlice(reflect.SliceOf(col.Source.Index(0).Type()), 0, 1) //create a new slice
				IsHere := func(item interface{}) (state bool) {
					state = false
					for k := 0; k < slice.Len(); k++ {
						if slice.Index(k).Interface() == item {
							state = true
						}
					}
					return
				}
				for i := 0; i < col.Source.Len(); i++ {
					if !IsHere(col.Source.Index(i).Interface()) {
						slice = reflect.Append(slice, col.Source.Index(i))
					}
				}

				col.Source = slice

			}
		},
	}

	c.Tasks = append(c.Tasks, t)       //add task
	c.Items = append(c.Items, empty{}) //add item of task

	return c
}
