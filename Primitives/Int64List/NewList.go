package Int64List


type query struct {
	source []int64
}

func NewList() query{
	var source []int64
	return query{
		source: source,
	}
}

func NewListFrom(s []int64) query{
	return query{
		source: s,
	}
}
