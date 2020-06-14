package StringList


type query struct {
	source []string
}

func NewList() query{
	var source []string
	return query{
		source: source,
	}
}

func NewListFrom(s []string) query{
	return query{
		source: s,
	}
}
