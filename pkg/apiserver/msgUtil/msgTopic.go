package msgutil

const (
	// RequestSchedule 请求调度
	NodeScheduleTopic = "nodeSchedule"

	EndpointUpdateTopic = "endpointUpdate"

	PodUpdateTopic = "podUpdate"

	ServiceUpdateTopic = "serviceUpdate"

	JobUpdateTopic = "jobUpdate"

	DnsUpdateTopic = "dnsUpdate"

	HostUpdateTopic = "hostUpdate"
)

func PodUpdateWithNode(node string) string {
	return PodUpdateTopic + "-" + node
}
