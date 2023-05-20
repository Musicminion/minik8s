package msgutil

const (
	// RequestSchedule 请求调度
	NodeSchedule = "nodeSchedule"

	EndpointUpdate = "endpointUpdate"

	PodUpdate = "podUpdate"

	ServiceUpdate = "serviceUpdate"

	JobUpdate = "jobUpdate"
)

func PodUpdateWithNode(node string) string {
	return PodUpdate + "-" + node
}
