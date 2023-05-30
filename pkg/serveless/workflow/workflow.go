package workflow

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/config"
	"miniK8s/util/executor"
	netrequest "miniK8s/util/netRequest"
	"miniK8s/util/stringutil"
	"net/http"
	"strconv"
)

type WorkflowController interface {
	Run()
}

type workflowController struct {
}

func NewWorkflowController() WorkflowController {
	return &workflowController{}
}

func GetAllWorkflowsFromAPIServer() ([]apiObject.WorkflowStore, error) {
	url := config.GetAPIServerURLPrefix() + config.GlobalWorkflowsURL

	allWorkflows := make([]apiObject.WorkflowStore, 0)

	code, err := netrequest.GetRequestByTarget(url, &allWorkflows, "data")

	if err != nil {
		return nil, err
	}

	if code != http.StatusOK {
		return nil, errors.New("get all workflows from apiserver failed")
	}

	return allWorkflows, nil
}

func (w *workflowController) routine() {
	allFlows, err := GetAllWorkflowsFromAPIServer()

	if err != nil {
		return
	}

	for _, flow := range allFlows {
		// 如果workflow的status不是空的，说明已经执行过了，就不再执行
		if flow.Status.Result != "" {
			continue
		}

		// 先更新workflow的phase为running
		if flow.Status.Phase == "" {
			w.WriteBackResultToServer(apiObject.WorkflowRunning, "", flow.GetNamespace(), flow.GetName())
		}

		// 如果workflow的某个func node无实例，进行创建，等待下一次的routine执行
		if !(w.checkWorkflowNode(flow)) {
			continue
		}

		go w.executeWorkflow(flow)
	}
}

func (w *workflowController) executeWorkflow(workflow apiObject.WorkflowStore) {
	nodeNameToNode := make(map[string]apiObject.WorkflowNode, 0)

	// 遍历workflow的所有node，构造一个nodeName到node的map
	for _, node := range workflow.Spec.WorkflowNodes {
		nodeNameToNode[node.Name] = node
	}

	curNodeName := workflow.Spec.EntryNodeName
	lastStepResult := workflow.Spec.EntryParams

	for {
		fmt.Println("curNodeName is ", curNodeName)
		fmt.Println("lastStepResult is ", lastStepResult)
		// 如果当前节点是空的，说明已经执行完了，就退出
		if curNodeName == "" {
			break
		}
		curNode, ok := nodeNameToNode[curNodeName]

		// 如果当前节点不存在，说明workflow配置有问题，就退出
		if !ok {
			break
		}

		// 如果是函数节点，就执行函数
		if curNode.Type == apiObject.WorkflowNodeTypeFunc {
			url := config.GetServelessServerURLPrefix() + "/" + curNode.FuncData.FuncNamespace + "/" + curNode.FuncData.FuncName
			resp, err := netrequest.PostString(url, lastStepResult)

			if err != nil {
				fmt.Println("post request failed + ", err.Error())
				break
			}

			if data, err := io.ReadAll(resp.Body); err == nil {
				defer resp.Body.Close()
				// 将data反序列化为map
				var result map[string]interface{}
				err := json.Unmarshal(data, &result)
				if err != nil {
					fmt.Println("unmarshal failed + ", err.Error())
					break
				}
				// 这里有可能在请求过程中pod被删除了，所以需要重新创建，workflow执行失败
				if result["data"] == nil {
					fmt.Println("result data is nil")
					break
				}
				lastStepResult = result["data"].(string)

			} else {
				fmt.Println("read resp body failed + ", err.Error())
				break
			}

			curNodeName = curNode.FuncData.NextNodeName

		} else if curNode.Type == apiObject.WorkflowNodeTypeChoice {
			res, err := w.CompareCheck(curNode.ChoiceData.CheckType, lastStepResult, curNode.ChoiceData.CompareValue, curNode.ChoiceData.CheckVarName)

			if err != nil {
				fmt.Println("compare check failed + ", err.Error())
				break
			}

			if res {
				curNodeName = curNode.ChoiceData.TrueNextNodeName
			} else {
				curNodeName = curNode.ChoiceData.FalseNextNodeName
			}
		}
	}

	w.WriteBackResultToServer(apiObject.WorkflowCompleted, lastStepResult, workflow.GetNamespace(), workflow.GetName())
}

func (w *workflowController) WriteBackResultToServer(phase string, result string, namespace string, name string) {
	statusURL := config.GetAPIServerURLPrefix() + config.WorkflowSpecStatusURL
	statusURL = stringutil.Replace(statusURL, config.URL_PARAM_NAMESPACE_PART, namespace)
	statusURL = stringutil.Replace(statusURL, config.URL_PARAM_NAME_PART, name)

	writeBackStatus := &apiObject.WorkflowStatus{
		Phase:  phase,
		Result: result,
	}

	code, _, err := netrequest.PutRequestByTarget(statusURL, writeBackStatus)

	if err != nil {
		fmt.Println("put request failed + ", err.Error())
		return
	}

	if code != http.StatusOK {
		fmt.Println("put request failed expected code 200, get " + strconv.Itoa(code))
		return
	}
}

// 检查当前的workflow的每个func node是否都存在pod实例，若不存在则创建并返回false
func (w *workflowController) checkWorkflowNode(workflow apiObject.WorkflowStore) bool {
	// 遍历workflow的所有node
	var ret = true
	for _, node := range workflow.Spec.WorkflowNodes {
		// 对于类型为func的node，检查node对应的function是否存在pod实例
		if node.Type == apiObject.WorkflowNodeTypeFunc {
			// 构造url
			url := config.GetServelessServerURLPrefix() + "/" + node.FuncData.FuncNamespace + "/" + node.FuncData.FuncName

			// 发送get请求
			var flag bool // 标记是否存在pod实例
			code, err := netrequest.GetRequestByTarget(url, &flag, "data")
			if err != nil || code != http.StatusOK || !flag {
				ret = false
			}
		}
	}
	return ret
}

func (w *workflowController) Run() {
	executor.Period(WorkflowController_Delay, WorkflowController_Waittime, w.routine, WorkflowController_ifLoop)
}

func (w *workflowController) CompareCheck(checkType apiObject.ChoiceCheckType, lastStepResult string, CompareValue string, varName string) (bool, error) {
	var result map[string]interface{}
	err := json.Unmarshal([]byte(lastStepResult), &result)

	if err != nil {
		return false, err
	}

	varValue, ok := result[varName]
	if !ok {
		return false, errors.New("varName not exist")
	}

	fmt.Println("varValue is ", varValue)
	var varValueFloat, CompareValueFloat float64
	var varValueString string

	if stringutil.ContainsString(apiObject.ChoiceCheckNumTypeList, string(checkType)) {
		// 如果type在ChoiceCheckNumTypeList之中
		// 把varValue转化为float，如果varValue不是float64，就报错
		if varValueFloat, ok = varValue.(float64); !ok {
			return false, errors.New("varValue is not float64")
		}
		// 把CompareValue转化为float64
		CompareValueFloat, err = strconv.ParseFloat(CompareValue, 64)
		if err != nil {
			return false, err
		}
	} else if stringutil.ContainsString(apiObject.ChoiceCheckStrTypeList, string(checkType)) {
		// 如果type在ChoiceCheckStringTypeList之中
		// 把varValue转化为string，如果varValue不是string，就报错
		varValueString, ok = varValue.(string)
		if !ok {
			return false, errors.New("varValue is not string")
		}
	} else {
		// 不存在的type
		return false, errors.New("checkType not exist")
	}

	switch checkType {
	// 以下为对num的比较
	case apiObject.ChoiceCheckTypeNumEqual:
		return varValueFloat == CompareValueFloat, nil

	case apiObject.ChoiceCheckTypeNumNotEqual:
		return varValueFloat != CompareValueFloat, nil

	case apiObject.ChoiceCheckTypeNumGreaterThan:
		return varValueFloat > CompareValueFloat, nil

	case apiObject.ChoiceCheckTypeNumLessThan:
		return varValueFloat < CompareValueFloat, nil

	case apiObject.ChoiceCheckTypeNumGreaterAndEqualThan:
		return varValueFloat >= CompareValueFloat, nil

	case apiObject.ChoiceCheckTypeNumLessAndEqualThan:
		return varValueFloat <= CompareValueFloat, nil

	// 以下为对string的比较
	case apiObject.ChoiceCheckTypeStrEqual:
		return varValueString == CompareValue, nil
	
	case apiObject.ChoiceCheckTypeStrNotEqual:
		return varValueString != CompareValue, nil

	case apiObject.ChoiceCheckTypeStrGreaterThan:
		return varValueString > CompareValue, nil

	case apiObject.ChoiceCheckTypeStrLessThan:
		return varValueString < CompareValue, nil

	case apiObject.ChoiceCheckTypeStrGreaterAndEqualThan:
		return varValueString >= CompareValue, nil

	case apiObject.ChoiceCheckTypeStrLessAndEqualThan:
		return varValueString <= CompareValue, nil
	}

	return false, errors.New("unknow checkType")
}
