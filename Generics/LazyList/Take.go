package LazyList


import "reflect"

func (c collection) Take(num int) collection { //it select records
	t := task{
		KindOfTask: "Void",
		Do: func() interface{} {
			return func(col *collection, g interface{}) {
				tk := g.(int)
				newSlice := reflect.MakeSlice(reflect.SliceOf(col.Source.Index(0).Type()), 0, 1)
				if tk > col.Source.Len() {
					tk = col.Source.Len()
				}
				for i := 0; i < tk; i++ {
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