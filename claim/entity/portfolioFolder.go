package entity

import (
	specentity "github.com/curltech/go-colla-biz/spec/entity"
	baseentity "github.com/curltech/go-colla-core/entity"
	"time"
)

type PortfolioFolder struct {
	specentity.InternalFixedActual `xorm:"extends"`
	BranchCompany                  string     `xorm:"varchar(255)" json:",omitempty"`
	BranchCompanyName              string     `xorm:"varchar(255)" json:",omitempty"`
	BusinessSpecCode               string     `xorm:"varchar(255)" json:",omitempty"`
	CloseDate                      *time.Time `json:",omitempty"`
	EditUserId                     uint64     `json:",omitempty"`
	EditUserName                   string     `xorm:"varchar(255)" json:",omitempty"`
	EstimatedAmount                float64    `json:",omitempty"`
	EventDate                      *time.Time `json:",omitempty"`
}

func (PortfolioFolder) TableName() string {
	return "clm_portfoliofolder"
}

func (PortfolioFolder) IdName() string {
	return baseentity.FieldName_Id
}
