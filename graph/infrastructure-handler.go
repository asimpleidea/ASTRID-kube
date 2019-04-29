package graph

import (
	informer "github.com/SunSince90/ASTRID-kube/informers"
	astrid_types "github.com/SunSince90/ASTRID-kube/types"
	log "github.com/sirupsen/logrus"
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type Infrastructure interface {
}

type InfrastructureHandler struct {
	clientset           kubernetes.Interface
	name                string
	labels              map[string]string
	deploymentsInformer cache.SharedIndexInformer
	servicesInformer    cache.SharedIndexInformer
	stopWatching        chan struct{}
}

func new(clientset kubernetes.Interface, namespace *core_v1.Namespace) (Infrastructure, error) {
	//	the handler
	inf := &InfrastructureHandler{
		name:      namespace.Name,
		labels:    namespace.Labels,
		clientset: clientset,
	}

	log.Infoln("Starting graph handler for graph", namespace.Name)

	deploymentsInformer := informer.New(astrid_types.Deployments, namespace.Name)
	deploymentsInformer.AddEventHandler(func(obj interface{}) {
		log.Infoln("new deployment!")
	}, nil, nil)

	/*deploymentsInformer := inf.getDeploymentsInformer()
	inf.deploymentsInformer = deploymentsInformer*/

	//servicesInformer := inf.getServicesInformer()
	//inf.servicesInformer = servicesInformer

	//stopWatching := make(chan struct{})

	//go deploymentsInformer.Run(stopWatching)
	//go servicesInformer.Run(stopWatching)
	//inf.stopWatching = stopWatching

	return inf, nil
}

func (handler *InfrastructureHandler) getServicesInformer() cache.SharedIndexInformer {
	//	Get the informer
	informer := cache.NewSharedIndexInformer(&cache.ListWatch{
		ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
			return handler.clientset.CoreV1().Services(handler.name).List(options)
		},
		WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
			return handler.clientset.CoreV1().Services(handler.name).Watch(options)
		},
	},
		&core_v1.Service{},
		0, //Skip resync
		cache.Indexers{},
	)

	//	Set the events
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			service := handler.getService(obj)
			if service != nil {
				//	do something about it
			}
		},
		UpdateFunc: func(old, new interface{}) {
		},
		DeleteFunc: func(obj interface{}) {
		},
	})
	return informer
}

func (handler *InfrastructureHandler) getService(obj interface{}) *core_v1.Service {
	//------------------------------------
	//	Try to get it
	//------------------------------------

	//	get the key
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		log.Errorln("Error while trying to parse a graph:", err)
		return nil
	}

	//	try to get the object
	_s, _, err := handler.deploymentsInformer.GetIndexer().GetByKey(key)
	if err != nil {
		log.Errorf("An error occurred: cannot find cache element with key %s from store %v", key, err)
		return nil
	}

	var service *core_v1.Service

	//	Get the namespace or try to recover it (this is a very improbable case, as we're doing this just for a new event).
	service, ok := _s.(*core_v1.Service)
	if !ok {
		service, ok = obj.(*core_v1.Service)
		if !ok {
			tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
			if !ok {
				log.Errorln("error decoding object, invalid type")
				return nil
			}
			service, ok = tombstone.Obj.(*core_v1.Service)
			if !ok {
				log.Errorln("error decoding object tombstone, invalid type")
				return nil
			}
			log.Infof("Recovered deleted object '%s' from tombstone", service.Name)
		}
	}

	//------------------------------------
	//	Add it
	//------------------------------------

	log.Infoln("Detected service with name:", service.Name)
	return service
}