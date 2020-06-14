package Float32List

type query struct {
	source []float32
}

func NewList() query{
	var source []float32
	return query{
		source: source,
	}
}

func NewListFrom(s []float32) query{
	return query{
		source: s,
	}
}
