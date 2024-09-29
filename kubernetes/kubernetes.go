package kubernetes

import (
	"context"
	"errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"

	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	eventsV1 "k8s.io/api/events/v1"
	networkingV1 "k8s.io/api/networking/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type AppService struct {
	Host  string
	Token string
}

// http client

func (k *AppService) Client() (*kubernetes.Clientset, error) {
	config := &rest.Config{
		Host:            k.Host,
		BearerToken:     k.Token,
		TLSClientConfig: rest.TLSClientConfig{Insecure: true},
	}
	return kubernetes.NewForConfig(config)
}

func (k *AppService) DynamicClient() (*dynamic.DynamicClient, error) {
	config := &rest.Config{
		Host:            k.Host,
		BearerToken:     k.Token,
		TLSClientConfig: rest.TLSClientConfig{Insecure: true},
	}
	return dynamic.NewForConfig(config)
}

// core

func (k *AppService) ListPods(namespace, labelSelector, fieldSelector string) (*coreV1.PodList, error) {
	client, err := k.Client()
	if err != nil {
		return nil, err
	}
	return client.CoreV1().Pods(namespace).List(
		context.Background(),
		metaV1.ListOptions{
			LabelSelector: labelSelector,
			FieldSelector: fieldSelector,
		},
	)
}

func (k *AppService) GetPod(name, namespace string) (*coreV1.Pod, error) {
	client, err := k.Client()
	if err != nil {
		return nil, err
	}
	return client.CoreV1().Pods(namespace).Get(context.Background(), name, metaV1.GetOptions{})
}

func (k *AppService) GetPodFromIp(ip, namespace, labelSelector, fieldSelector string) (*coreV1.Pod, error) {
	if ip == "" {
		return nil, errors.New("the ip parameter cannot be empty")
	}
	filterItems, listErr := k.ListPods(namespace, labelSelector, fieldSelector)
	if listErr != nil {
		return nil, listErr
	}
	for _, item := range filterItems.Items {
		if item.Status.PodIP == ip {
			return &item, nil
		}
	}
	return nil, errors.New("a nonexistent pod ip or filter parameter is incorrect")
}

func (k *AppService) GetPodFromName(name, namespace, labelSelector, fieldSelector string) (*coreV1.Pod, error) {
	if name == "" {
		return nil, errors.New("the name parameter cannot be empty")
	}
	filterItems, listErr := k.ListPods(namespace, labelSelector, fieldSelector)
	if listErr != nil {
		return nil, listErr
	}
	for _, item := range filterItems.Items {
		if item.Name == name {
			return &item, nil
		}
	}
	return nil, errors.New("a nonexistent pod name or filter parameter is incorrect")
}

func (k *AppService) UpdatePod(namespace string, pod *coreV1.Pod, dryRun bool) (*coreV1.Pod, error) {
	client, err := k.Client()
	if err != nil {
		return nil, err
	}
	var DryRun []string
	if dryRun {
		DryRun = []string{"All"}
	}
	return client.CoreV1().Pods(namespace).Update(context.Background(), pod, metaV1.UpdateOptions{DryRun: DryRun})
}

func (k *AppService) PatchPod(name, namespace string, MergePatchTypeData []byte, dryRun bool) (*coreV1.Pod, error) {
	client, err := k.Client()
	if err != nil {
		return nil, err
	}
	var DryRun []string
	if dryRun {
		DryRun = []string{"All"}
	}
	return client.CoreV1().Pods(namespace).Patch(
		context.Background(), name, types.MergePatchType, MergePatchTypeData, metaV1.PatchOptions{DryRun: DryRun},
	)
}

func (k *AppService) ServicesList(namespace, labelSelector, fieldSelector string) (*coreV1.ServiceList, error) {
	client, err := k.Client()
	if err != nil {
		return nil, err
	}
	return client.CoreV1().Services(namespace).List(context.Background(),
		metaV1.ListOptions{
			LabelSelector: labelSelector,
			FieldSelector: fieldSelector,
		},
	)
}

// apps

func (k *AppService) GetDeployments(name, namespace string) (*appsV1.Deployment, error) {
	client, err := k.Client()
	if err != nil {
		return nil, err
	}
	return client.AppsV1().Deployments(namespace).Get(context.Background(), name, metaV1.GetOptions{})
}

func (k *AppService) GetStatefulSets(name, namespace string) (*appsV1.StatefulSet, error) {
	client, err := k.Client()
	if err != nil {
		return nil, err
	}
	return client.AppsV1().StatefulSets(namespace).Get(context.Background(), name, metaV1.GetOptions{})
}

func (k *AppService) GetDaemonSetsSets(name, namespace string) (*appsV1.DaemonSet, error) {
	client, err := k.Client()
	if err != nil {
		return nil, err
	}
	return client.AppsV1().DaemonSets(namespace).Get(context.Background(), name, metaV1.GetOptions{})
}

func (k *AppService) UpdateDeployments(namespace string, deployments *appsV1.Deployment, dryRun bool) (*appsV1.Deployment, error) {
	client, err := k.Client()
	if err != nil {
		return nil, err
	}
	var DryRun []string
	if dryRun {
		DryRun = []string{"All"}
	}
	return client.AppsV1().Deployments(namespace).Update(
		context.Background(), deployments, metaV1.UpdateOptions{DryRun: DryRun},
	)
}

// PatchDeployments name: DeploymentsName, namespace: namespace, MergePatchTypeData: []byte(`{"metadata":{"labels":{"tag":"test"}}}`)
func (k *AppService) PatchDeployments(name, namespace string, MergePatchTypeData []byte, dryRun bool) (*appsV1.Deployment, error) {
	client, err := k.Client()
	if err != nil {
		return nil, err
	}
	var DryRun []string
	if dryRun {
		DryRun = []string{"All"}
	}
	return client.AppsV1().Deployments(namespace).Patch(
		context.Background(), name, types.MergePatchType, MergePatchTypeData, metaV1.PatchOptions{DryRun: DryRun},
	)
}

func (k *AppService) ListDeployments(namespace, labelSelector, fieldSelector string) (*appsV1.DeploymentList, error) {
	client, err := k.Client()
	if err != nil {
		return nil, err
	}
	return client.AppsV1().Deployments(namespace).List(context.Background(),
		metaV1.ListOptions{
			LabelSelector: labelSelector,
			FieldSelector: fieldSelector,
		},
	)
}

func (k *AppService) ListStatefulSets(namespace, labelSelector, fieldSelector string) (*appsV1.StatefulSetList, error) {
	client, err := k.Client()
	if err != nil {
		return nil, err
	}
	return client.AppsV1().StatefulSets(namespace).List(context.Background(),
		metaV1.ListOptions{
			LabelSelector: labelSelector,
			FieldSelector: fieldSelector,
		},
	)
}

func (k *AppService) ListDaemonSets(namespace, labelSelector, fieldSelector string) (*appsV1.DaemonSetList, error) {
	client, err := k.Client()
	if err != nil {
		return nil, err
	}
	return client.AppsV1().DaemonSets(namespace).List(context.Background(),
		metaV1.ListOptions{
			LabelSelector: labelSelector,
			FieldSelector: fieldSelector,
		},
	)
}

// networking

func (k *AppService) ListIngresses(namespace, labelSelector, fieldSelector string) (*networkingV1.IngressList, error) {
	client, err := k.Client()
	if err != nil {
		return nil, err
	}
	return client.NetworkingV1().Ingresses(namespace).List(context.Background(),
		metaV1.ListOptions{
			LabelSelector: labelSelector,
			FieldSelector: fieldSelector,
		},
	)
}

// events

func (k *AppService) ListEvents(namespace, labelSelector, fieldSelector string) (*eventsV1.EventList, error) {
	client, err := k.Client()
	if err != nil {
		return nil, err
	}
	return client.EventsV1().Events(namespace).List(context.Background(),
		metaV1.ListOptions{
			LabelSelector: labelSelector,
			FieldSelector: fieldSelector,
		},
	)
}

// watch

func (k *AppService) WatchDeployments() (chan struct{}, error) {
	client, err := k.Client()
	if err != nil {
		return nil, err
	}
	factory := informers.NewSharedInformerFactory(client, 0)
	informer := factory.Apps().V1().Deployments().Informer()
	_, err = informer.AddEventHandler(NewEventHandler())
	if err != nil {
		return nil, err
	}
	stopper := make(chan struct{}, 2)
	informer.Run(stopper)
	return stopper, nil
}

func (k *AppService) WatchStatefulSets() (chan struct{}, error) {
	client, err := k.Client()
	if err != nil {
		return nil, err
	}
	factory := informers.NewSharedInformerFactory(client, 0)
	informer := factory.Apps().V1().StatefulSets().Informer()
	_, err = informer.AddEventHandler(NewEventHandler())
	if err != nil {
		return nil, err
	}
	stopper := make(chan struct{}, 2)
	informer.Run(stopper)
	return stopper, nil
}

func (k *AppService) WatchDaemonSets() (chan struct{}, error) {
	client, err := k.Client()
	if err != nil {
		return nil, err
	}
	factory := informers.NewSharedInformerFactory(client, 0)
	informer := factory.Apps().V1().DaemonSets().Informer()
	_, err = informer.AddEventHandler(NewEventHandler())
	if err != nil {
		return nil, err
	}
	stopper := make(chan struct{}, 2)
	informer.Run(stopper)
	return stopper, nil
}

func (k *AppService) WatchPods() (chan struct{}, error) {
	client, err := k.Client()
	if err != nil {
		return nil, err
	}
	factory := informers.NewSharedInformerFactory(client, 0)
	informer := factory.Core().V1().Pods().Informer()
	_, err = informer.AddEventHandler(NewEventHandler())
	if err != nil {
		return nil, err
	}
	stopper := make(chan struct{}, 2)
	informer.Run(stopper)
	return stopper, nil
}

func (k *AppService) WatchPodsRestart(handler *EventHandlerPodsRestart) (chan struct{}, error) {
	client, err := k.Client()
	if err != nil {
		return nil, err
	}
	factory := informers.NewSharedInformerFactory(client, 0)
	informer := factory.Core().V1().Pods().Informer()
	_, err = informer.AddEventHandler(handler)
	if err != nil {
		return nil, err
	}
	stopper := make(chan struct{}, 2)
	informer.Run(stopper)
	return stopper, nil
}
