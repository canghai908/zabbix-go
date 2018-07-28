package zabbix

import (
	"github.com/AlekSi/reflector"
)

type HistoryItem struct {
	ItemId string `json:"itemid"`
	Clock  string `json:"clock"`
	Value  string `json:"value"`
	ns     string `json:"ns"`
}

type HistoryItems []HistoryItem

func (api *API) HistoryGet(params Params) (res HistoryItems, err error) {
	if _, present := params["output"]; !present {
		params["output"] = "extend"
	}
	if _, presentl := params["limit"]; !presentl {
		params["limit"] = "100"
	}
	if _, presenth := params["history"]; !presenth {
		params["history"] = "0"
	}
	response, err := api.CallWithError("history.get", params)
	if err != nil {
		return
	}

	reflector.MapsToStructs2(response.Result.([]interface{}), &res, reflector.Strconv, "json")
	return
}
