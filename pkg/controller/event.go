package controller

type eventType string

const (
	addHello eventType = "addEcho"
)

type event struct {
	eventType      eventType
	oldObj, newObj interface{}
}
