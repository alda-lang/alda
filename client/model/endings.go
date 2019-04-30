package model

type EndingRange struct {
	First int32
	Last  int32
}

type Endings struct {
	Ranges []EndingRange
}
