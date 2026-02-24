package identityObject

import (
	"fmt"

	"github.com/towgo/towgo/dao/basedboperat"
	"github.com/towgo/towgo/module/accountcenter/jurisdictionObject"

	"log"
)

func FindIdentityIdsByJurisdictionCode(code string) []interface{} {
	var identityIds []interface{}
	identityIds = append(identityIds, 0)
	// 根据code查出角色
	var jurisdistions []jurisdictionObject.Jurisdiction
	err := basedboperat.QueryScan(&jurisdistions, nil, "code = ?", code)
	if err != nil {
		log.Println(fmt.Errorf("FindIdentityIdsByJurisdictionCode: %w", err))
		return identityIds
	}
	// 解析权限ID
	var jurisdistionIds []interface{}
	for _, j := range jurisdistions {
		jurisdistionIds = append(jurisdistionIds, j.ID)
	}
	jurisdistionIds = append(jurisdistionIds, 0)
	// 根据权限ID查出角色ID
	var identities []IdentityJurisdictions
	err = basedboperat.QueryScan(&identities, &basedboperat.ListSimple{Table: IdentityJurisdictions{}.TableName(), In: map[string][]interface{}{"jurisdiction_id": jurisdistionIds}}, nil)
	if err != nil {
		log.Println(fmt.Errorf("FindIdentityIdsByJurisdictionCode: %w", err))
		return identityIds
	}
	for _, j := range identities {
		identityIds = append(identityIds, j.IdentityId)
	}
	return identityIds
}
