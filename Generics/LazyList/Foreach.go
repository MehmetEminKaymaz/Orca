package LazyList


import "reflect"

func (c collection) Foreach(do func(x interface{}) (y interface{})) collection { //foreach
	t := task{
		KindOfTask: "Func", //its about parameter (do func(x interface{}) (y interface{}))
		Do: func() interface{} {
			return func(col *collection, fdo func(x interface{}) (y interface{})) {
				for i := 0; i < col.Source.Len(); i++ {
					col.Source.Index(i).Set(reflect.ValueOf(fdo(col.Source.Index(i).Interface())))
				}
			}
		},
	}

	c.Tasks = append(c.Tasks, t)  //add task
	c.Items = append(c.Items, do) //add item of task

	return c
}
