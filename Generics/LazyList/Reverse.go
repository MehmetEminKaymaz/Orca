package LazyList

import "reflect"

func (c collection) Reverse() collection { //reverse the collection
	t := task{
		KindOfTask: "Void2",
		Do: func() interface{} {
			return func(col *collection) {
				newSlice := reflect.MakeSlice(reflect.SliceOf(col.Source.Index(0).Type()), 0, 1)
				for i := col.Source.Len() - 1; i >= 0; i-- {
					newSlice = reflect.Append(newSlice, col.Source.Index(i))
				}
				col.Source = newSlice
			}
		},
	}

	c.Tasks = append(c.Tasks, t)
	c.Items = append(c.Items, empty{})

	return c
}