package LazyList


func (c collection) ElementAt(index int) collection { //it returns element by index value
	t := task{
		KindOfTask: "Void",
		Do: func() interface{} {
			return func(col *collection, i interface{}) {
				col.Source = col.Source.Index(i.(int))
			}
		},
	}

	c.Tasks = append(c.Tasks, t)
	c.Items = append(c.Items, index)

	return c

}