package monitor

import "testing"

func TestStartProcM(t *testing.T) {
	netlink := new(Netlink)
	netlink.StartProcM()
	for {
		netlink.Receive()
	}
}
