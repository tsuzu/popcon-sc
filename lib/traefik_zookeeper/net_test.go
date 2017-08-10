package traefikZookeeper

import "testing"

func TestDefaultIPAddress(t *testing.T) {
	addr, err := DefaultIPAddress()

	if err != nil {
		t.Logf("DefaultIPAddress() error(%s)", err.Error())
	}

	if len(addr) == 0 {
		t.Error("the length of the address is 0")
		t.FailNow()
	}

	t.Log("Address:", addr)
}

func TestIPAddressFromIface(t *testing.T) {
	addr, err := IPAddressFromIface("enp3s0")

	if err != nil {
		t.Errorf("IPAddressFromIface() error(%s)", err.Error())
	}

	if len(addr) == 0 {
		t.Error("the length of the address is 0")
		t.FailNow()
	}

	t.Log("Address: ", addr)
}
