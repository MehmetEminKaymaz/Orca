package LazyList


import "reflect"

func (c collection) Skip(num int) collection { //skip elements (until num) in the collection
	t := task{
		KindOfTask: "Void",
		Do: func() interface{} {
			return func(col *collection, skp interface{}) {
				newSlice := reflect.MakeSlice(reflect.SliceOf(col.Source.Index(0).Type()), 0, 1)
				for i := skp.(int); i < col.Source.Len(); i++ {
					newSlice = reflect.Append(newSlice, col.Source.Index(i))
				}

				col.Source = newSlice
			}
		},
	}

	c.Tasks = append(c.Tasks, t)
	c.Items = append(c.Items, num)

	return c
}