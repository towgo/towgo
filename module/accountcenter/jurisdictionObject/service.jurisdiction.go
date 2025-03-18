package jurisdictionObject

import (
	"github.com/towgo/towgo/dao/basedboperat"
)

func TreeList(j Jurisdictions) []Jurisdictions {

	var jurisdictions Jurisdictions
	var jurisdictionss []Jurisdictions
	var list basedboperat.List
	list.Limit = -1
	list.And = map[string][]interface{}{
		"fid": []interface{}{j.ID},
	}
	basedboperat.ListScan(&list, jurisdictions, &jurisdictionss)

	for k, v := range jurisdictionss {
		jurisdictionss[k].Childs = TreeList(v)
	}
	return jurisdictionss
}

func AllChildList(j Jurisdiction) []Jurisdiction {
	var allChildJurisdictionlist []Jurisdiction
	var jurisdictionList []Jurisdiction
	basedboperat.SqlQueryScan(&jurisdictionList, "select * from jurisdictions where fid = ?", j.ID)
	allChildJurisdictionlist = append(allChildJurisdictionlist, jurisdictionList...)
	for _, v := range jurisdictionList {
		allChildJurisdictionlist = append(allChildJurisdictionlist, AllChildList(v)...)
	}
	return allChildJurisdictionlist
}
