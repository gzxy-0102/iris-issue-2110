package common

type ActionEnum uint

type StateEnum uint

const (
	CREATE ActionEnum = iota
	UPDATE
	SELECT
	DELETE

	UNKNOW
)

const (
	DEVELOPING StateEnum = iota
	ONLINE
	OFFFLINE
)
