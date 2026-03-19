package dto

// OrganizationInfo 组织信息值对象
type OrganizationInfo struct {
	OrgID     int    `json:"orgId"`
	OrgCode   string `json:"orgCode"`
	OrgName   string `json:"orgName"`
	OrgNameJc string `json:"orgNameJc"` // 组织简称
	FullName  string `json:"fullName"`
	ParentID  *int   `json:"parentId,omitempty"`
}

// UserInfo 用户信息值对象
type UserInfo struct {
	UserID   int32  `json:"userId"`
	UserName string `json:"userName"`
	PoliceNo string `json:"policeNo"`
	OrgID    int32  `json:"orgId"`
}

// BWCInfo 执法记录仪信息值对象
type BWCInfo struct {
	ID              int32  `json:"id"`
	BWCNo           string `json:"no"`
	BWCName         string `json:"name"`
	RequisitionerId int32  `json:"requisitionerId"`
}

// StorageSiteInfo 存储站点信息值对象
type StorageSiteInfo struct {
	ID              int32  `json:"id"`              // 主键ID
	StorageSiteNo   string `json:"storageSiteNo"`   // 存储站点编号
	StorageSiteName string `json:"storageSiteName"` // 存储站点名称
	StorageSiteIP   string `json:"storageSiteIp"`   // IP地址
	StorageSiteURL  string `json:"storageSiteUrl"`  // 播放地址(HTTP)
	OpenStatus      int32  `json:"openStatus"`      // 启用状态(0禁用/1启用)
	AuthKey         string `json:"authKey"`         // 认证密钥
}

// EnforcementTypeInfo 执法类型信息值对象
type EnforcementTypeInfo struct {
	ID                  int64  `json:"id"`                  // 主键ID
	EnforcementTypeCode string `json:"enforcementTypeCode"` // 执法类型编码
	EnforcementTypeName string `json:"enforcementTypeName"` // 执法类型名称
	EnforcementTypeDesc string `json:"enforcementTypeDesc"` // 执法类型描述
	EnforcementTypePath string `json:"enforcementTypePath"` // 执法类型路径
	ParentId            int64  `json:"parentId"`            // 父级ID
	Source              string `json:"source"`              // 来源
	Sort                int32  `json:"sort"`                // 排序
}
