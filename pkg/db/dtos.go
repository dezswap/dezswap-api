package db

type LastIdLimitCondition struct {
	LastId    uint
	Limit     int
	DescOrder bool
}
