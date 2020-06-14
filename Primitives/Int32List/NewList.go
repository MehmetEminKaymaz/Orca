package Int32List

type query struct {
	source []int32
}

func NewList() query{
	var source []int32
	return query{
		source: source,
	}
}

func NewListFrom(s []int32) query{
	return query{
		source: s,
	}
}
