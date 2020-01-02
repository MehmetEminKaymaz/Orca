package LazyList


func (c collection) First() collection { //it returns the first element of the collection
	t := task{
		KindOfTask: "Void2",
		Do: func() interface{} {
			return func(col *collection) {
				col.Source = col.Source.Index(0)
			}
		},
	}

	c.Tasks = append(c.Tasks, t)
	c.Items = append(c.Items, empty{})
	return c
}
