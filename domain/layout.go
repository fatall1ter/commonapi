package domain

import (
	"strings"
)

// Layout is like a object:
// mall, retail, airport, railway station, transport company, store, vehicle etc...
type Layout struct {
	LayoutID string      `json:"layout_id"`
	Title    string      `json:"title,omitempty"`
	Kind     string      `json:"kind,omitempty"`
	Owner    CRMCustomer `json:"owner,omitempty"`
	IsActive bool        `json:"is_active,omitempty"`
	Stat     LayoutStat  `json:"stat,omitempty"`
}

type LayoutStat struct {
	TotalDevices         int64 `json:"total_devices,omitempty"`
	TotalEnters          int64 `json:"total_enters,omitempty"`
	TotalZones           int64 `json:"total_zones,omitempty"`
	TotalStores          int64 `json:"total_stores,omitempty"`
	TotalRenters         int64 `json:"total_renters,omitempty"`
	TotalOpenedIncidents int64 `json:"total_opened_incidents,omitempty"`
}

type Layouts []Layout

// ACL return only accessible list ids
func (los Layouts) ACL(listIDs string) (string, int) {
	const sep string = ","
	slList := strings.Split(listIDs, sep)
	slRes := make([]string, 0)
	if listIDs == "" || listIDs == "*" {
		for _, has := range los {
			slRes = append(slRes, has.LayoutID)
		}
		return strings.Join(slRes, sep), len(slRes)
	}
	for _, chk := range slList {
		for _, has := range los {
			if has.LayoutID == chk {
				slRes = append(slRes, chk)
			}
		}
	}
	return strings.Join(slRes, sep), len(slRes)
}

// CRMCustomer properties of customers from CRM system
type CRMCustomer struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}
