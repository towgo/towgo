/*
 * @Author       : lvyitao lvyitao@fanhaninfo.com
 * @Date         : 2024-06-03 13:30:34
 * @LastEditTime : 2024-06-03 13:35:04
 */
package jurisdictionObject

type IdentityJurisdictionsObject struct {
	ID             int64
	IdentityId     int64 `json:"identity_id"`
	JurisdictionId int64 `json:"jurisdiction_id"`
}

func (IdentityJurisdictionsObject) TableName() string {
	return "identitys_jurisdictions"
}
