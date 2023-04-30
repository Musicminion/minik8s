package proxy

import (
    "context"
    "flag"
    "fmt"
    "log"
    "net"
    "net/http"
    "os"
    "sync"
    "time"
    "miniK8s/pkg/"
    // "k8s.io/apimachinery/pkg/util/wait"
    // "k8s.io/client-go/kubernetes"
    // "k8s.io/client-go/rest"
    // "k8s.io/client-go/tools/cache"
    // "k8s.io/client-go/util/workqueue"
)

var (
    serviceCIDR = flag.String("service-cidr", "10.244.0.0/16", "Service CIDR")
)

type KubeProxy struct {
}
func main() {
    flag.Parse()

    // 初始化 Kubernetes REST 配置
    config, err := rest.InClusterConfig()
    if err != nil {
        log.Fatalf("Failed to load Kubernetes config: %v", err)
    }

    // 建立 Kubernetes 客户端
    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        log.Fatalf("Failed to create Kubernetes client: %v", err)
    }

    // 创建 Kube-Proxy Server 实例
    proxyServer := &ProxyServer{
        Clientset: clientset,
        ServiceCIDR: *serviceCIDR,
    }

    // 启动 Kube-Proxy Server
    if err := proxyServer.Start(); err != nil {
        log.Fatalf("Failed to start Kube-Proxy server: %v", err)
    }
}

// ProxyServer 定义 Kube-Proxy Server 结构
type ProxyServer struct {
    Clientset   *kubernetes.Clientset // Kubernetes 客户端
    ServiceCIDR string                // Service CIDR

    // endpointsIndexer 存储 endpoint 列表的 indexer
    endpointsIndexer cache.Indexer
    // endpointsSynced 用于同步 endpointsIndexer 的标记
    endpointsSynced cache.InformerSynced
    // endpointsQueue 存储 endpoint 变更事件的队列
    endpointsQueue workqueue.RateLimitingInterface
}

// Start 启动 Kube-Proxy Server
func (s *ProxyServer) Start() error {
    // 创建 endpointsIndexer 和 endpointsSynced
    s.endpointsIndexer, s.endpointsSynced = cache.NewIndexerInformer(
        // ListWatch 函数返回一个用于监视 endpoint 变化的 ListerWatcher，这里使用了默认的 ListWatch 实现。
        cache.NewListWatchFromClient(s.Clientset.CoreV1().RESTClient(), "endpoints", "", nil),
        // Controller 需要一个缓存对象，用于存储 endpoint 列表，这里使用了 cache.NewIndexer() 函数创建一个新的 indexer。
        cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{}),
        // resyncPeriod 是指定多久重新获取一次 endpoint 列表，默认为 0（不重新获取）。
        time.Second * 30,
        // endpointUpdateHandler 是用于处理 endpoint 变化事件的回调函数。
        cache.ResourceEventHandlerFuncs{
            AddFunc:    s.endpointUpdateHandler,
            UpdateFunc: s.endpointUpdateHandler,
            DeleteFunc: s.endpointUpdateHandler,
        },
    )

    // 创建 endpointsQueue
    s.endpointsQueue = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

    // 启动 endpointsIndexer 和 endpointsSynced
    go s.endpointsIndexer.Run(make(chan struct{}))
    if !cache.WaitForCacheSync(make(chan struct{}), s.endpointsSynced) {
        return fmt.Errorf("Failed to sync endpoints cache")
    }

    // 启动 endpointsQueue 处理循环
    go wait.Until(s.processEndpointsQueue, time.Second, make(chan struct{}))

    // 启动 HTTP 服务器
    http.HandleFunc("/", s.handleRequest)
    port := os.Getenv("KUBE_PROXY_PORT")
    if port == "" {
        port = "8080"
    }
    log.Printf("Listening on :%s...\n", port)
    if err := http.ListenAndServe(":"+port, nil); err != nil {
        return err
    }

    return nil
}

// handleRequest 处理 HTTP 请求
func (s *ProxyServer) handleRequest(w http.ResponseWriter, r *http.Request) {
    log.Printf("Received request: %s %s %s\n", r.Method, r.Host, r.URL.Path)

    // 解析请求 URL
    serviceName := r.URL.Query().Get("service")
    if serviceName == "" {
        http.Error(w, "service parameter missing", http.StatusBadRequest)
        return
    }

    // 获取 Service IP
    serviceIP, err := s.getServiceIP(serviceName)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    // 修改请求 Host
    r.Host = serviceIP

    // 发送请求到目标地址
    proxy := &httputil.ReverseProxy{Director: func(req *http.Request) {
        req.URL.Scheme = "http"
        req.URL.Host = serviceIP
    }}
    proxy.ServeHTTP(w, r)
}

// endpointUpdateHandler 处理 endpoint 变化事件
func (s *ProxyServer) endpointUpdateHandler(obj interface{}) {
    // 将 endpoint 变化事件加入 endpointsQueue 中
    objKey, err := cache.MetaNamespaceKeyFunc(obj)
    if err != nil {
        log.Printf("Error processing endpoint update: %v\n", err)
        return
    }
    s.endpointsQueue.Add(objKey)
}

// getServiceIP 获取 Service IP
func (s *ProxyServer) getServiceIP(serviceName string) (string, error) {
    // 获取 endpoint 列表
    endpoints, err := s.getEndpoints(serviceName)
    if err != nil {
        return "", err
    }

    // 根据 endpoint 列表计算出 Service IP
    serviceIP, err := s.calculateServiceIP(serviceName, endpoints)
    if err != nil {
        return "", err
    }

    return serviceIP, nil
}

// getEndpoints 获取 endpoint 列表
func (s *ProxyServer) getEndpoints(serviceName string) ([]string, error) {
    // 根据 service 名称获取 endpoints 列表
    endpoints, err := s.endpointsIndexer.ByIndex("byService", cache.MetaNamespaceKey(serviceName, ""))
    if err != nil {
        return nil, err
    }

    var endpointList []string
    for _, obj := range endpoints {
        endpoint, ok := obj.(*corev1.Endpoints)
        if !ok {
            return nil, fmt.Errorf("Invalid endpoint object: %#v", obj)
        }

        for _, subset := range endpoint.Subsets {
            for _, address := range subset.Addresses {
                endpointList = append(endpointList, address.IP)
            }
        }
    }

    return endpointList, nil
}

// calculateServiceIP 根据 endpoint 列表计算出 Service IP
func (s *ProxyServer) calculateServiceIP(serviceName string, endpoints []string) (string, error) {
    // 计算 Service IP
    ip := net.ParseIP("0.0.0.0")
    for _, endpoint := range endpoints {
        endpointIP := net.ParseIP(endpoint)
        if endpointIP == nil {
            return "", fmt.Errorf("Invalid endpoint IP address: %s", endpoint)
        }
        for i := 0; i < 16; i++ {
            ip[i] |= endpointIP[i]
        }
    }

    // 检查是否在 Service CIDR 内
    _, serviceCIDR, err := net.ParseCIDR(s.ServiceCIDR)
    if err != nil {
        return "", fmt.Errorf("Invalid service CIDR: %v", err)
    }
    if !serviceCIDR.Contains(ip) {
        return "", fmt.Errorf("Service IP %s is not in CIDR %s", ip, serviceCIDR)
    }

    return ip.String(), nil
}

// processEndpointsQueue 处理 endpointsQueue 中的 endpoint 变化事件
func (s *ProxyServer) processEndpointsQueue() {
    for {
        key, quit := s.endpointsQueue.Get()
        if quit {
            return
        }

        if err := s.updateIptables(key.(string)); err != nil {
            log.Printf("Failed to update iptables for %q: %v", key, err)
            s.endpointsQueue.AddRateLimited(key)
        } else {
            s.endpointsQueue.Forget(key)
        }

        s.endpointsQueue.Done(key)
    }
}

func (s *ProxyServer) updateIptables(key string) error{

}
