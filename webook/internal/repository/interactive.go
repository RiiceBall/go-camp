package repository

import (
	"context"
	intrv1 "webook/api/proto/gen/intr/v1"
	"webook/interactive/domain"
)

type InteractiveRepository interface {
	Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error)
}

type GRPCInteractiveRepository struct {
	grpcClient intrv1.InteractiveServiceClient
}

func NewInteractiveRepository(client intrv1.InteractiveServiceClient) *GRPCInteractiveRepository {
	return &GRPCInteractiveRepository{grpcClient: client}
}

func (repo *GRPCInteractiveRepository) Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error) {
	// 调用 gRPC 服务获取数据
	resp, err := repo.grpcClient.Get(ctx, &intrv1.GetRequest{
		Biz:   biz,
		BizId: bizId,
	})
	if err != nil {
		return domain.Interactive{}, err
	}
	return domain.Interactive{
		Biz:        "article",
		BizId:      bizId,
		ReadCnt:    resp.Intr.ReadCnt,
		LikeCnt:    resp.Intr.LikeCnt,
		CollectCnt: resp.Intr.CollectCnt,
		Liked:      resp.Intr.Liked,
		Collected:  resp.Intr.Collected,
	}, nil
}
