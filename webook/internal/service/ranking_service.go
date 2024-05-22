package service

import (
	"context"
	"math"
	"time"
	intrv1 "webook/api/proto/gen/intr/v1"
	"webook/internal/domain"
	"webook/internal/repository"

	"github.com/ecodeclub/ekit/queue"
	"github.com/ecodeclub/ekit/slice"
)

type RankingService interface {
	// TopN 前 100 的
	TopN(ctx context.Context) error
	GetTopN(ctx context.Context) ([]domain.Article, error)
}

type BatchRankingService struct {
	// 用来取点赞数
	intrSvc intrv1.InteractiveServiceClient

	// 用来查找文章
	artSvc ArticleService

	batchSize int
	scoreFunc func(likeCnt int64, utime time.Time) float64
	n         int

	rr repository.RankingRepository
}

func NewBatchRankingService(intrSvc intrv1.InteractiveServiceClient, artSvc ArticleService) RankingService {
	return &BatchRankingService{
		intrSvc:   intrSvc,
		artSvc:    artSvc,
		batchSize: 100,
		n:         100,
		scoreFunc: func(likeCnt int64, utime time.Time) float64 {
			// 时间
			duration := time.Since(utime).Seconds()
			return float64(likeCnt-1) / math.Pow(duration+2, 1.5)
		},
	}
}

func (rs *BatchRankingService) GetTopN(ctx context.Context) ([]domain.Article, error) {
	return rs.rr.GetTopN(ctx)
}

func (rs *BatchRankingService) TopN(ctx context.Context) error {
	arts, err := rs.topN(ctx)
	if err != nil {
		return err
	}
	// 最终是要放到缓存里面的
	// 存到缓存里面
	return rs.rr.ReplaceTopN(ctx, arts)
}

func (rs *BatchRankingService) topN(ctx context.Context) ([]domain.Article, error) {
	offset := 0
	start := time.Now()
	ddl := start.Add(-7 * 24 * time.Hour)

	type Score struct {
		score float64
		art   domain.Article
	}
	topN := queue.NewPriorityQueue[Score](rs.n,
		func(src Score, dst Score) int {
			if src.score > dst.score {
				return 1
			} else if src.score == dst.score {
				return 0
			} else {
				return -1
			}
		})

	for {
		// 取数据
		arts, err := rs.artSvc.ListPub(ctx, start, offset, rs.batchSize)
		if err != nil {
			return nil, err
		}
		//if len(arts) == 0 {
		//	break
		//}
		ids := slice.Map(arts, func(idx int, art domain.Article) int64 {
			return art.Id
		})
		// 取点赞数
		intrResp, err := rs.intrSvc.GetByIds(ctx, &intrv1.GetByIdsRequest{
			Biz: "article", Ids: ids,
		})
		if err != nil {
			return nil, err
		}
		intrMap := intrResp.Intrs
		for _, art := range arts {
			intr := intrMap[art.Id]
			//intr, ok := intrMap[art.Id]
			//if !ok {
			//	continue
			//}
			score := rs.scoreFunc(intr.LikeCnt, art.Utime)
			ele := Score{
				score: score,
				art:   art,
			}
			err = topN.Enqueue(ele)
			if err == queue.ErrOutOfCapacity {
				// 这个也是满了
				// 拿出最小的元素
				minEle, _ := topN.Dequeue()
				if minEle.score < score {
					_ = topN.Enqueue(ele)
				} else {
					_ = topN.Enqueue(minEle)
				}
			}
		}
		offset = offset + len(arts)
		// 没有取够一批，我们就直接中断执行
		// 没有下一批了
		if len(arts) < rs.batchSize ||
			// 这个是一个优化
			arts[len(arts)-1].Utime.Before(ddl) {
			break
		}
	}

	// 这边 topN 里面就是最终结果
	res := make([]domain.Article, topN.Len())
	for i := topN.Len() - 1; i >= 0; i-- {
		ele, _ := topN.Dequeue()
		res[i] = ele.art
	}
	return res, nil
}
