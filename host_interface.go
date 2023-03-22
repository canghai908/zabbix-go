package zabbix

import (
	"github.com/canghai908/reflector"
)

type (
	InterfaceType int
)

const (
	Agent InterfaceType = 1
	SNMP  InterfaceType = 2
	IPMI  InterfaceType = 3
	JMX   InterfaceType = 4
)

// https://www.zabbix.com/documentation/2.2/manual/appendix/api/hostinterface/definitions
type HostInterface struct {
	HostId      string        `json:"hostid"`
	InterfaceId string        `json:"interfaceid"`
	DNS         string        `json:"dns"`
	IP          string        `json:"ip"`
	Main        int           `json:"main"`
	Port        string        `json:"port"`
	Type        InterfaceType `json:"type"`
	UseIP       int           `json:"useip"`
}

type HostInterfaces []HostInterface

func (api *API) HostInterfacesGet(params Params) (res HostInterfaces, err error) {
	if _, present := params["output"]; !present {
		params["output"] = "extend"
	}
	if _, presentl := params["limit"]; !presentl {
		params["limit"] = "100"
	}

	response, err := api.CallWithError("hostinterface.get", params)
	if err != nil {
		return
	}

	reflector.MapsToStructs2(response.Result.([]interface{}), &res, reflector.Strconv, "json")
	return
}
