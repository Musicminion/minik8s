package server

import (
	"fmt"
	"io"
	"math/rand"
	"miniK8s/pkg/k8log"
	netrequest "miniK8s/util/netRequest"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (s *server) UpdateCountRecord(name, namespace string) {
	// TODO
	// 在s.callRecord中找到对应的记录，然后更新
	// key := namespace + "/" + name
	// record, ok := s.funcController.CallRecord[key]

	// if !ok {
	// 	// 如果没有找到，说明是第一次调用
	// 	record = &LaunchRecord{
	// 		StartTime:     time.Now(),
	// 		EndTime:       time.Now().Add(time.Duration(5) * time.Minute),
	// 		FuncName:      name,
	// 		FuncNamespace: namespace,
	// 		FuncCallTime:  1,
	// 	}
	// 	s.callRecord[key] = record
	// } else {
	// 	cur := time.Now()
	// 	// 检查当前时间是否在record的EndTime之前
	// 	if cur.Before(record.EndTime) {
	// 		// 在EndTime之前，说明是同一个周期内的调用
	// 		record.FuncCallTime++
	// 	} else {
	// 		// 不在EndTime之前，说明是一个新的周期
	// 		record.StartTime = cur
	// 		record.EndTime = cur.Add(time.Duration(5) * time.Minute)
	// 		record.FuncCallTime = 1
	// 	}
	// }

}

// handleFuncRequest
// /:funcNamespace/:funcName
func (s *server) handleFuncRequest(c *gin.Context) {
	// 解析请求参数里面的funcNamespace和funcName
	funcNamespace := c.Param("funcNamespace")
	funcName := c.Param("funcName")

	k8log.InfoLog("serveless", "func: "+funcNamespace+"/"+funcName+" is called")
	if funcNamespace == "" || funcName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "funcNamespace or funcName is empty",
		})
		return
	}

	// 查询routeTable，找到对应的pod的ip地址
	key := funcNamespace + "/" + funcName
	podIPs, ok := s.routeTable[key]

	if !ok {
		s.funcController.ScaleUp(funcName, funcNamespace, 2)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "The function you call has no pod running, maybe creating, please try again later",
		})
		return
	}

	// 随机选择一个pod的ip地址
	if len(podIPs) == 0 {
		s.funcController.ScaleUp(funcName, funcNamespace, 2)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "The function you call has no pod running, maybe creating, please try again later",
		})
		return
	}

	// 产生一个随机的数，大小为0-(len(podIPs)-1)之间
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randID := r.Intn(len(podIPs))
	podIP := podIPs[randID]

	// 将请求转发到pod上
	// 读取请求的body
	body, err := c.GetRawData()

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "read request body error",
		})
		return
	}

	// 把请求的body打印出来
	fmt.Println(string(body))

	// 将请求转发到pod上
	url := "http://" + podIP + ":18080"

	resp, err := http.Post(url, "application/json", c.Request.Body)

	if err != nil {
		s.funcController.ScaleUp(funcNamespace, funcName, 2)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "forward request to pod error, " + err.Error(),
		})
		return
	}

	defer resp.Body.Close()

	// 读取pod的响应
	// var respJson interface{}

	// // 打印输出resp的body
	// fmt.Println(resp.Body)

	// if err := json.NewDecoder(resp.Body).Decode(&respJson); err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{
	// 		"message": "forward request to pod error, " + err.Error(),
	// 	})
	// 	return
	// }

	var respPtr *http.Response

	if respPtr, err = netrequest.PostString(url, string(body)); err == nil {
		var data []byte
		if data, err = io.ReadAll(respPtr.Body); err == nil {
			defer respPtr.Body.Close()
			c.JSON(http.StatusOK, gin.H{
				"data": string(data),
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "forward request to pod error 1, " + err.Error(),
			})
		}
	} else {
		s.funcController.ScaleUp(funcNamespace, funcName, 2)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "forward request to pod error 2, " + err.Error(),
		})
	}

	// 对被请求的function，添加callrecord
	err = s.funcController.AddCallRecord(funcName, funcNamespace)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "add call record error, " + err.Error(),
		})
	}

	// code, respBody, err := netrequest.PostRequestByTarget(url, body)

	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{
	// 		"message": "forward request to pod error, " + err.Error(),
	// 	})
	// 	return
	// }

	// if code != http.StatusOK {
	// 	c.JSON(http.StatusBadRequest, gin.H{
	// 		"message": "forward request to pod error, code is not 200",
	// 	})
	// 	return
	// }

	// bodyBytes, err := json.Marshal(respBody)

	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{
	// 		"message": "forward request to pod error" + err.Error(),
	// 	})
	// 	return
	// }

	// c.JSON(http.StatusOK, gin.H{
	// 	"data": string(bodyBytes),
	// })
}

func (s *server) checkFunction(c *gin.Context) {
	// 解析请求参数里面的funcNamespace和funcName
	funcNamespace := c.Param("funcNamespace")
	funcName := c.Param("funcName")

	k8log.InfoLog("serveless", "checkout func: "+funcNamespace+"/"+funcName)
	if funcNamespace == "" || funcName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "funcNamespace or funcName is empty",
		})
		return
	}

	// 判断function是否存在pod实例
	ips := s.routeTable[funcNamespace + "/" + funcName]
	if  len(ips) == 0{
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "function has no pod running",
			"data": false,
		})
		// 不存在，需要创建实例
		s.funcController.ScaleUp(funcName, funcNamespace, 2)
		return
	}

	// 存在，添加callrecord
	err := s.funcController.AddCallRecord(funcName, funcNamespace)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "add call record error, " + err.Error(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data":    true,
	})
}

// func checkAndCreate(funcNamespace, funcName string) {
// 	// 尝试通过api server获取到function的信息
// 	url := config.GetAPIServerURLPrefix() + config.FunctionSpecURL
// 	url = stringutil.Replace(url, config.URL_PARAM_NAMESPACE_PART, funcNamespace)
// 	url = stringutil.Replace(url, config.URL_PARAM_NAME_PART, funcName)

// 	funcObj := apiObject.Function{}
// 	code, err := netrequest.GetRequestByTarget(url, &funcObj, "data")

// 	if err != nil {
// 		return
// 	}

// 	if code != http.StatusOK {
// 		return
// 	}

// 	// 尝试通过api server获取到function的对应的replicaset的信息
// 	url = config.GetAPIServerURLPrefix() + config.ReplicaSetSpecURL
// 	url = stringutil.Replace(url, config.URL_PARAM_NAMESPACE_PART, funcNamespace)
// 	url = stringutil.Replace(url, config.URL_PARAM_NAME_PART, funcName)

// 	replicaSetObj := apiObject.ReplicaSet{}
// 	code, err = netrequest.GetRequestByTarget(url, &replicaSetObj, "data")

// 	if err != nil {
// 		return
// 	}

// 	if code != http.StatusOK {
// 		// 这说明还没有创建replicaset，需要创建

// 	} else {
// 		//
// 		return
// 	}

// }
