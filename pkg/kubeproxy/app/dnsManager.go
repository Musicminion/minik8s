package proxy

type DnsManager interface {
	// Run 运行plegManager
	Run()
}

type dnsManager struct {

}

// 创建PlegManager的时候，必须要传递一个statusManager，以及PlegChannel
func NewDnsManager() DnsManager {
	return &dnsManager{
	}
}

func (p *dnsManager) Run() {
	// TODO
}