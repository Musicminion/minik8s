package function

import (
	"bytes"
	"errors"
	"fmt"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/config"
	"miniK8s/pkg/k8log"
	"miniK8s/util/executor"
	netrequest "miniK8s/util/netRequest"
	"miniK8s/util/stringutil"
	"net/http"
	"strconv"
	"time"
)

type LaunchRecord struct {
	StartTime     time.Time
	EndTime       time.Time
	FuncName      string
	FuncNamespace string
	FuncCallTime  int
}

type FuncController interface {
	Run()

	GetFuncRecord(funcName, funcNamespace string) *LaunchRecord
	AddCallRecord(funcName, funcNamespace string) error

	ScaleUp(funcName, funcNamespace string, num int) error
	ScaleDown(funcName, funcNamespace string) error
}

type funcController struct {
	cache      map[string]*apiObject.Function
	CallRecord map[string]*LaunchRecord // key: namespace/funcName
}

func NewFuncController() FuncController {
	return &funcController{
		cache:      make(map[string]*apiObject.Function),
		CallRecord: make(map[string]*LaunchRecord),
	}
}

func (c *funcController) getAllFunc() ([]apiObject.Function, error) {
	url := config.GetAPIServerURLPrefix() + config.GlobalFunctionsURL

	allFuncs := make([]apiObject.Function, 0)

	code, err := netrequest.GetRequestByTarget(url, &allFuncs, "data")

	if err != nil {
		return nil, err
	}

	if code != http.StatusOK {
		return nil, errors.New("get all functions from apiserver failed, not 200")
	}

	return allFuncs, nil
}

func (c *funcController) routine() {
	res, err := c.getAllFunc()
	if err != nil {
		return
	}

	remoteFuncs := make(map[string]bool)
	for id, f := range res {
		remoteFuncs[f.Metadata.UUID] = true
		// 检查f是否在cache中
		if _, ok := c.cache[f.Metadata.UUID]; !ok {
			// 如果不在cache中，说明是新的function，需要创建
			// 【TODO】
			c.cache[f.Metadata.UUID] = &res[id]
			c.CallRecord[f.Metadata.Namespace+"/"+f.Metadata.Name] = &LaunchRecord{
				FuncName:      f.Metadata.Name,
				FuncNamespace: f.Metadata.Namespace,
				StartTime:     time.Now(),
				EndTime:       time.Now().Add(time.Duration(1) * time.Minute),
			}
			c.CreateFunction(&res[id])

		} else {
			// 检查c.CallRecord是否过期
			if c.CallRecord[f.Metadata.Namespace+"/"+f.Metadata.Name] != nil {
				if time.Now().After(c.CallRecord[f.Metadata.Namespace+"/"+f.Metadata.Name].EndTime) {

					if c.CallRecord[f.Metadata.Namespace+"/"+f.Metadata.Name].FuncCallTime == 0 {
						// 缩容
						c.ScaleDown(f.Metadata.Name, f.Metadata.Namespace)
					} else if c.CallRecord[f.Metadata.Namespace+"/"+f.Metadata.Name].FuncCallTime > 100 {
						newSize := 2 + c.CallRecord[f.Metadata.Namespace+"/"+f.Metadata.Name].FuncCallTime/100
						// 扩容
						c.ScaleUp(f.Metadata.Name, f.Metadata.Namespace, newSize)
					}

					// 过期了，需要重置
					c.CallRecord[f.Metadata.Namespace+"/"+f.Metadata.Name].FuncCallTime = 0
					c.CallRecord[f.Metadata.Namespace+"/"+f.Metadata.Name].StartTime = time.Now()
					c.CallRecord[f.Metadata.Namespace+"/"+f.Metadata.Name].EndTime = time.Now().Add(time.Duration(1) * time.Minute)
				}
			}

			// 如果在cache中，说明是已经存在的function，需要检查是否需要更新
			// 【TODO】
			if !c.ComplareTwoFunc(c.cache[f.Metadata.UUID], &res[id]) {
				c.cache[f.Metadata.UUID] = &res[id]
				c.UpdateFunction(&res[id])
				fmt.Println("update function")
			} else {
				c.cache[f.Metadata.UUID] = &res[id]
			}
		}
	}

	// 检查cache中的function是否在remoteFuncs中
	for uuid, f := range c.cache {
		if _, ok := remoteFuncs[uuid]; !ok {
			// 如果不在remoteFuncs中，说明需要删除
			//
			c.DeleteFunction(f)
			delete(c.cache, uuid)
			delete(c.CallRecord, f.Metadata.Namespace+"/"+f.Metadata.Name)
		}
	}

	// 这样保证这边的缓存和apiserver中的缓存一致
}

// 比较两个function是否相同
func (c *funcController) ComplareTwoFunc(old *apiObject.Function, new *apiObject.Function) bool {
	// 【TODO】
	if old.Spec.UserUploadFilePath != new.Spec.UserUploadFilePath {
		fmt.Println("old.Spec.UserUploadFilePath != new.Spec.UserUploadFilePath")
		return false
	}
	if len(old.Spec.UserUploadFile) != len(new.Spec.UserUploadFile) {
		fmt.Println("len(old.Spec.UserUploadFile) != len(new.Spec.UserUploadFile)")
		return false
	}
	if !bytes.Equal(old.Spec.UserUploadFile, new.Spec.UserUploadFile) {
		fmt.Println("file content is not the same")
		return false
	}
	return true
}

func (c *funcController) Run() {
	executor.Period(FuncControllerUpdateDelay, FuncControllerUpdateFrequency, c.routine, FuncControllerUpdateLoop)
}

func (c *funcController) GetFuncRecord(funcName, funcNamespace string) *LaunchRecord {
	return c.CallRecord[funcNamespace+"/"+funcName]
}

func (c *funcController) AddCallRecord(funcName, funcNamespace string) error {

	if c.CallRecord[funcNamespace+"/"+funcName] != nil {
		c.CallRecord[funcNamespace+"/"+funcName].FuncCallTime++
		k8log.DebugLog("func call time: ", strconv.Itoa(c.CallRecord[funcNamespace+"/"+funcName].FuncCallTime))
	} else {
		c.CallRecord[funcNamespace+"/"+funcName] = &LaunchRecord{
			FuncName:      funcName,
			FuncNamespace: funcNamespace,
			StartTime:     time.Now(),
			EndTime:       time.Now().Add(time.Duration(1) * time.Minute),
			FuncCallTime:  1,
		}
	}
	return nil
}

func (c *funcController) ScaleDown(funcName, funcNamespace string) error {
	fmt.Println("scale down start")
	// 【TODO】
	url := config.GetAPIServerURLPrefix() + config.ReplicaSetSpecURL
	url = stringutil.Replace(url, config.URL_PARAM_NAMESPACE_PART, funcNamespace)
	url = stringutil.Replace(url, config.URL_PARAM_NAME_PART, funcName)

	replica := &apiObject.ReplicaSetStore{}
	code, err := netrequest.GetRequestByTarget(url, replica, "data")

	if err != nil {
		return err
	}

	if code != http.StatusOK {
		return errors.New("get function from apiserver failed, not 200")
	}

	if replica.Spec.Replicas > 0 {
		replica.Spec.Replicas = replica.Spec.Replicas / 2
	}

	code, _, err = netrequest.PutRequestByTarget(url, replica)

	if err != nil {
		fmt.Println("put function from apiserver failed" + err.Error())
		return err
	}

	if code != http.StatusOK {
		fmt.Println("put function from apiserver failed, not 200")
		return errors.New("put function from apiserver failed, not 200")
	}

	fmt.Println("scale down end")
	return nil
}

// 【TODO】
func (c *funcController) ScaleUp(funcName, funcNamespace string, num int) error {
	fmt.Println("funcName scale up to ", num)
	// 【TODO】
	url := config.GetAPIServerURLPrefix() + config.ReplicaSetSpecURL
	url = stringutil.Replace(url, config.URL_PARAM_NAMESPACE_PART, funcNamespace)
	url = stringutil.Replace(url, config.URL_PARAM_NAME_PART, funcName)

	replica := &apiObject.ReplicaSetStore{}
	code, err := netrequest.GetRequestByTarget(url, replica, "data")

	if err != nil {
		return err
	}

	if code != http.StatusOK {
		return errors.New("get function from apiserver failed, not 200")
	}

	if num > 0 {
		replica.Spec.Replicas = num
	}

	code, _, err = netrequest.PutRequestByTarget(url, replica)

	if err != nil {
		return err
	}

	if code != http.StatusOK {
		return errors.New("put function from apiserver failed, not 200")
	}

	return nil
}
