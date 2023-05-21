package allcontollers

// import (
// 	"miniK8s/pkg/entity"
// 	"miniK8s/pkg/listwatcher"
// )

// type ServiceController struct {
// 	// serviceCacheManager *ServiceCacheManager
// 	serviceMap  map[string]entity.ServiceWithEndpoints
// 	lw          *listwatcher.Listwatcher
// 	stopChannel chan struct{}
// }

// func NewServiceController(lw *listwatcher.Listwatcher) *ServiceController {
// 	return &ServiceController{
// 		serviceMap:  make(map[string]entity.ServiceWithEndpoints),
// 		lw:          lw,
// 		stopChannel: make(chan struct{}),
// 	}
// }

// func (sc *ServiceController) Run() {
// }
