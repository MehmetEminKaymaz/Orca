package LazyList


import "reflect"

func (c collection) RemoveAt(index int) collection { //it removes element at index
	t := task{
		KindOfTask: "Void",
		Do: func() interface{} {
			return func(col *collection, i interface{}) {
				newSlice := reflect.MakeSlice(reflect.SliceOf(col.Source.Index(0).Type()), 0, 1)

				for i := 0; i < col.Source.Len(); i++ {
					if index != i {
						newSlice = reflect.Append(newSlice, col.Source.Index(i))
					}

				}
				col.Source = newSlice
			}
		},
	}

	c.Tasks = append(c.Tasks, t)     //add task
	c.Items = append(c.Items, index) //add item of task

	return c
}