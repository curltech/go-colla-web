package service

import (
	"github.com/curltech/go-colla-biz/claim/entity"
	"github.com/curltech/go-colla-core/container"
	"github.com/curltech/go-colla-core/service"
	"github.com/curltech/go-colla-core/util/message"
)

/**
同步表结构，服务继承基本服务的方法
*/
type PortfolioFolderService struct {
	service.OrmBaseService
}

var portfolioFolderService = &PortfolioFolderService{}

func GetSessionInstanceService() *PortfolioFolderService {
	return portfolioFolderService
}

var seqname = "seq_actual"

func (this *PortfolioFolderService) GetSeqName() string {
	return seqname
}

func (this *PortfolioFolderService) NewEntity(data []byte) (interface{}, error) {
	entity := &entity.PortfolioFolder{}
	if data == nil {
		return entity, nil
	}
	err := message.Unmarshal(data, entity)
	if err != nil {
		return nil, err
	}

	return entity, err
}

func (this *PortfolioFolderService) NewEntities(data []byte) (interface{}, error) {
	entities := make([]*entity.PortfolioFolder, 0)
	if data == nil {
		return &entities, nil
	}
	err := message.Unmarshal(data, &entities)
	if err != nil {
		return nil, err
	}

	return &entities, err
}

func init() {
	service.GetSession().Sync(new(entity.PortfolioFolder))
	portfolioFolderService.OrmBaseService.GetSeqName = portfolioFolderService.GetSeqName
	portfolioFolderService.OrmBaseService.FactNewEntity = portfolioFolderService.NewEntity
	portfolioFolderService.OrmBaseService.FactNewEntities = portfolioFolderService.NewEntities
	service.RegistSeq(seqname, 0)
	container.RegistService("portfolioFolder", portfolioFolderService)
}
