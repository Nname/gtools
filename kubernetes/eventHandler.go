package kubernetes

import (
	"fmt"
	"github.com/gogf/gf/v2/encoding/gjson"
	coreV1 "k8s.io/api/core/v1"
)

type EventHandler struct {
}

func NewEventHandler() *EventHandler {
	return &EventHandler{}
}

func (e *EventHandler) OnAdd(obj interface{}) {
	//
	fmt.Println("OnAdd...")
}
func (e *EventHandler) OnUpdate(oldObj, newObj interface{}) {
	//
	fmt.Println("OnUpdate...")
}
func (e *EventHandler) OnDelete(obj interface{}) {
	//
	fmt.Println("OnDelete...")
}

type EventHandlerPodsRestart struct {
}

func NewEventHandlerPodsRestart() *EventHandlerPodsRestart {
	return &EventHandlerPodsRestart{}
}

func (e *EventHandlerPodsRestart) OnAdd(obj interface{}) {
	//
}

func (e *EventHandlerPodsRestart) OnUpdate(oldObj, newObj interface{}) {
	oldObjJson := gjson.New(oldObj.(*coreV1.Pod))
	newObjJson := gjson.New(newObj.(*coreV1.Pod))
	if oldObjJson.Get("status.containerStatuses.0.restartCount").Int()+1 == newObjJson.Get("status.containerStatuses.0.restartCount").Int() {
		fmt.Println(fmt.Sprintf(`{"namespace": %s,"container": %s,"pod": %s, "restartCount": %s, "uid": %s, "creationTimestamp": %s}`,
			newObjJson.Get("metadata.namespace"),
			newObjJson.Get("status.containerStatuses.0.name"),
			newObjJson.Get("metadata.name"),
			newObjJson.Get("status.containerStatuses.0.restartCount"),
			newObjJson.Get("metadata.uid"),
			newObjJson.Get("metadata.creationTimestamp.Time"),
		))
		fmt.Println(">>>>>>>>>>>>>>>>>>>>>>")
	}
}

func (e *EventHandlerPodsRestart) OnDelete(obj interface{}) {
	//
}
