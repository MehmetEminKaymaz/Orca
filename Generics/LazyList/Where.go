package LazyList


import "reflect"

func (c collection) Where(ok func(x interface{}) bool) collection { //returns data according to a given condition
	t := task{
		KindOfTask: "Bool",
		Do: func() interface{} {
			return func(col *collection, okk func(x interface{}) bool) {
				newSlice := reflect.MakeSlice(reflect.SliceOf(col.Source.Index(0).Type()), 0, 1)

				for i := 0; i < col.Source.Len(); i++ {
					if ok(col.Source.Index(i).Interface()) {
						newSlice = reflect.Append(newSlice, col.Source.Index(i))
					}
				}

				col.Source = newSlice
			}
		},
	}

	c.Tasks = append(c.Tasks, t)
	c.Items = append(c.Items, ok)

	return c

}
