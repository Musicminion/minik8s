package server

import (
	"fmt"
	"io"
	"math/rand"
	netrequest "miniK8s/util/netRequest"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// handleFuncRequest
// /:funcNamespace/:funcName
func (s *server) handleFuncRequest(c *gin.Context) {
	// 解析请求参数里面的funcNamespace和funcName
	funcNamespace := c.Param("funcNamespace")
	funcName := c.Param("funcName")

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
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "funcNamespace or funcName is not exist, maybe creating",
		})

		return
	}

	// 随机选择一个pod的ip地址
	if len(podIPs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "funcNamespace or funcName is not exist, maybe creating",
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
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "forward request to pod error 2, " + err.Error(),
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

// func checkAndCreate(funcNamespace, funcName string) {
// 	// 尝试通过api server获取到function的信息
// 	url := config.API_Server_URL_Prefix + config.FunctionSpecURL
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
// 	url = config.API_Server_URL_Prefix + config.ReplicaSetSpecURL
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
