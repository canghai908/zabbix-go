package zabbix_test

import (
	"testing"

	. "github.com/canghai908/zabbix-go"
)

func TestHostsGet_List(t *testing.T) {
	api := getAPI(t)

	hosts, err := api.HostsGet(Params{"output": "extend", "limit": 1})
	if err != nil {
		t.Fatal(err)
	}
	if hosts == nil {
		t.Fatal("hosts is nil")
	}
}

func TestHostGroupsGet_List(t *testing.T) {
	api := getAPI(t)

	groups, err := api.HostGroupsGet(Params{"output": "extend", "limit": 1})
	if err != nil {
		t.Fatal(err)
	}
	if groups == nil {
		t.Fatal("groups is nil")
	}
}

