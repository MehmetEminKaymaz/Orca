package LazyList


func (c collection) Last() collection { //it returns the last element of the collection
	t := task{
		KindOfTask: "Void2",
		Do: func() interface{} {
			return func(col *collection) {
				col.Source = col.Source.Index(col.Source.Len() - 1)
			}
		},
	}

	c.Tasks = append(c.Tasks, t)
	c.Items = append(c.Items, empty{})

	return c
}