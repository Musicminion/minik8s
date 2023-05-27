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
	url := config.API_Server_URL_Prefix + config.GlobalWorkflowsURL

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
			url := "http://" + config.API_Server_IP + ":28080" + curNode.FuncData.FuncNamespace + "/" + curNode.FuncData.FuncName
			resp, err := netrequest.PostString(url, lastStepResult)

			if err != nil {
				fmt.Println("post request failed + ", err.Error())
				break
			}

			if data, err := io.ReadAll(resp.Body); err != nil {
				defer resp.Body.Close()
				lastStepResult = string(data)
			} else {
				fmt.Println("read resp body failed + ", err.Error())
				break
			}

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

	switch checkType {
	case apiObject.ChoiceCheckTypeEqual:
		// 把varValue转化为int，然后和curNode.ChoiceData.CheckValue比较
		// 如果相等，就执行curNode.ChoiceData.EqualNodeName
		varValueInt := varValue.(int)
		CompareValueInt, err := strconv.Atoi(CompareValue)

		if err != nil {
			return false, err
		}

		// CompareValue也转化为int
		if varValueInt == CompareValueInt {
			return true, nil
		} else {
			return false, nil
		}

	case apiObject.ChoiceCheckTypeNotEqual:
		varValueInt := varValue.(int)
		CompareValueInt, err := strconv.Atoi(CompareValue)

		if err != nil {
			return false, err
		}

		// CompareValue也转化为int
		if varValueInt != CompareValueInt {
			return true, nil
		} else {
			return false, nil
		}

	case apiObject.ChoiceCheckTypeNumGreaterThen:
		varValueInt := varValue.(int)
		CompareValueInt, err := strconv.Atoi(CompareValue)

		if err != nil {
			return false, err
		}

		// CompareValue也转化为int
		if varValueInt > CompareValueInt {
			return true, nil
		} else {
			return false, nil
		}

	case apiObject.ChoiceCheckTypeNumLessThen:

		varValueInt := varValue.(int)
		CompareValueInt, err := strconv.Atoi(CompareValue)

		if err != nil {
			return false, err
		}

		// CompareValue也转化为int
		if varValueInt < CompareValueInt {
			return true, nil
		} else {
			return false, nil
		}

	case apiObject.ChoiceCheckTypeNumGreaterAndEqualThen:
		varValueInt := varValue.(int)
		CompareValueInt, err := strconv.Atoi(CompareValue)

		if err != nil {
			return false, err
		}

		// CompareValue也转化为int
		if varValueInt >= CompareValueInt {
			return true, nil
		} else {
			return false, nil
		}

	case apiObject.ChoiceCheckTypeNumLessAndEqualThen:
		varValueInt := varValue.(int)
		CompareValueInt, err := strconv.Atoi(CompareValue)

		if err != nil {
			return false, err
		}

		// CompareValue也转化为int
		if varValueInt <= CompareValueInt {
			return true, nil
		} else {
			return false, nil
		}

	case apiObject.ChoiceCheckTypeStrGreaterThen:
		varValueInt := varValue.(string)

		// CompareValue也转化为int
		if varValueInt > CompareValue {
			return true, nil
		} else {
			return false, nil
		}

	case apiObject.ChoiceCheckTypeStrLessThen:

		varValueInt := varValue.(string)

		// CompareValue也转化为int
		if varValueInt < CompareValue {
			return true, nil
		} else {
			return false, nil
		}

	case apiObject.ChoiceCheckTypeStrGreaterAndEqualThen:
		varValueInt := varValue.(string)

		// CompareValue也转化为int
		if varValueInt >= CompareValue {
			return true, nil
		} else {
			return false, nil
		}
	case apiObject.ChoiceCheckTypeStrLessAndEqualThen:
		varValueInt := varValue.(string)

		// CompareValue也转化为int
		if varValueInt <= CompareValue {
			return true, nil
		} else {
			return false, nil
		}
	}

	return false, errors.New("unknow checkType")
}

// ChoiceCheckTypeEqual                  ChoiceCheckType = "equal"
// ChoiceCheckTypeNotEqual               ChoiceCheckType = "notEqual"
// ChoiceCheckTypeNumGreaterThen         ChoiceCheckType = "numGreaterThen"
// ChoiceCheckTypeNumLessThen            ChoiceCheckType = "numLessThen"
// ChoiceCheckTypeNumGreaterAndEqualThen ChoiceCheckType = "numGreaterAndEqualThen"
// ChoiceCheckTypeNumLessAndEqualThen    ChoiceCheckType = "numLessAndEqualThen"

// ChoiceCheckTypeStrGreaterThen         ChoiceCheckType = "strGreaterThen"
// ChoiceCheckTypeStrLessThen            ChoiceCheckType = "strLessThen"
// ChoiceCheckTypeStrGreaterAndEqualThen ChoiceCheckType = "strGreaterAndEqualThen"
// ChoiceCheckTypeStrLessAndEqualThen    ChoiceCheckType = "strLessAndEqualThen"