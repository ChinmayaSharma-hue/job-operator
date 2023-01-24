package controller

type eventType string

const (
	addHello    eventType = "addHello"
	addHelloJob eventType = "addHelloJob"
)

type event struct {
	eventType                     eventType
	oldObj, newObj                interface{}
	custom_resource, job_resource interface{}
}
