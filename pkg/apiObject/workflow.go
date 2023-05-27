package apiObject

type Workflow struct {
	Basic `json:",inline" yaml:",inline"`
	Spec  WorkflowSpec `json:"spec" yaml:"spec"`
}

type WorkflowNodeType string

const (
	WorkflowNodeTypeFunc   WorkflowNodeType = "func"
	WorkflowNodeTypeChoice WorkflowNodeType = "choice"
)

type WorkflowFuncData struct {
	FuncName      string `json:"funcName" yaml:"funcName"`
	FuncNamespace string `json:"funcNamespace" yaml:"funcNamespace"`
	NextNodeName  string `json:"nextNodeName" yaml:"nextNodeName"`
}

type ChoiceCheckType string

const (
	ChoiceCheckTypeEqual                  ChoiceCheckType = "equal"
	ChoiceCheckTypeNotEqual               ChoiceCheckType = "notEqual"
	ChoiceCheckTypeNumGreaterThen         ChoiceCheckType = "numGreaterThen"
	ChoiceCheckTypeNumLessThen            ChoiceCheckType = "numLessThen"
	ChoiceCheckTypeNumGreaterAndEqualThen ChoiceCheckType = "numGreaterAndEqualThen"
	ChoiceCheckTypeNumLessAndEqualThen    ChoiceCheckType = "numLessAndEqualThen"

	ChoiceCheckTypeStrGreaterThen         ChoiceCheckType = "strGreaterThen"
	ChoiceCheckTypeStrLessThen            ChoiceCheckType = "strLessThen"
	ChoiceCheckTypeStrGreaterAndEqualThen ChoiceCheckType = "strGreaterAndEqualThen"
	ChoiceCheckTypeStrLessAndEqualThen    ChoiceCheckType = "strLessAndEqualThen"
)

type WorkflowChoiceData struct {
	TrueNextNodeName  string `json:"trueNextNodeName" yaml:"trueNextNodeName"`
	FalseNextNodeName string `json:"falseNextNodeName" yaml:"falseNextNodeName"`

	CheckType    ChoiceCheckType `json:"checkType" yaml:"checkType"`
	CheckVarName string          `json:"checkVarName" yaml:"checkVarName"`
	// 需要保证能够从上一个结果中获取到,填写json的key

	CompareValue string `json:"compareValue" yaml:"compareValue"` // 需要比较的值(无论是数字还是字符串，都需要转化为字符串)
}

type WorkflowNode struct {
	Name       string             `json:"name" yaml:"name"`
	Type       WorkflowNodeType   `json:"type" yaml:"type"`
	FuncData   WorkflowFuncData   `json:"funcData" yaml:"funcData"`
	ChoiceData WorkflowChoiceData `json:"choiceData" yaml:"choiceData"`
}

type WorkflowSpec struct {
	EntryParams   string         `json:"entryParams" yaml:"entryParams"`
	EntryNodeName string         `json:"entryNodeName" yaml:"entryNodeName"`
	WorkflowNodes []WorkflowNode `json:"workflowNodes" yaml:"workflowNodes"`
}

type WorkflowStatus struct {
	Result string `json:"result" yaml:"result"`
}

type WorkflowStore struct {
	Basic  `json:",inline" yaml:",inline"`
	Spec   WorkflowSpec   `json:"spec" yaml:"spec"`
	Status WorkflowStatus `json:"status" yaml:"status"`
}

// 定义Workflow转化为WorkflowStore的函数
func (w *Workflow) ToWorkflowStore() *WorkflowStore {
	// 创建一个Status是空的WorkflowStore
	return &WorkflowStore{
		Basic:  w.Basic,
		Spec:   w.Spec,
		Status: WorkflowStatus{},
	}
}

// 定义WorkflowStore转化为Workflow的函数
func (w *WorkflowStore) ToWorkflow() *Workflow {
	// 创建一个Status是空的WorkflowStore
	return &Workflow{
		Basic: w.Basic,
		Spec:  w.Spec,
	}
}

// 定义获取name、namespace的函数
func (w *Workflow) GetName() string {
	return w.Metadata.Name
}

func (w *Workflow) GetNamespace() string {
	return w.Metadata.Namespace
}

// 定义获取name、namespace的函数
func (w *WorkflowStore) GetName() string {
	return w.Metadata.Name
}

func (w *WorkflowStore) GetNamespace() string {
	return w.Metadata.Namespace
}
