package validator

import (
	"context"
	"time"
	"webook/pkg/logger"
	"webook/pkg/migrator"
	"webook/pkg/migrator/events"

	"github.com/ecodeclub/ekit/slice"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

type Validator[T migrator.Entity] struct {
	// 你数据迁移，是不是肯定有
	base   *gorm.DB
	target *gorm.DB

	l         logger.LoggerV1
	producer  events.Producer
	direction string
	batchSize int
}

func (v *Validator[T]) Validate(ctx context.Context) error {
	//err := v.validateBaseToTarget(ctx)
	//if err != nil {
	//	return err
	//}
	//return v.validateTargetToBase(ctx)

	var eg errgroup.Group
	eg.Go(func() error {
		return v.validateBaseToTarget(ctx)
	})
	eg.Go(func() error {
		return v.validateTargetToBase(ctx)
	})
	return eg.Wait()
}

func (v *Validator[T]) validateBaseToTarget(ctx context.Context) error {
	offset := -1
	for {
		// 进来就检测负载
		offset++
		var src T
		err := v.base.WithContext(ctx).Order("id").Offset(offset).First(&src).Error
		if err == gorm.ErrRecordNotFound {
			// 这个就是咩有数据
			return nil
		}
		if err != nil {
			// 查询出错了
			v.l.Error("base -> target 查询 base 失败", logger.Error(err))
			// 在这里，
			continue
		}

		// 这边就是正常情况
		var dst T
		err = v.target.WithContext(ctx).
			Where("id = ?", src.ID()).
			First(&dst).Error
		switch err {
		case gorm.ErrRecordNotFound:
			// target 没有
			// 丢一条消息到 Kafka 上
			v.notify(src.ID(), events.InconsistentEventTypeTargetMissing)
		case nil:
			equal := src.CompareTo(dst)
			if !equal {
				// 要丢一条消息到 Kafka 上
				v.notify(src.ID(), events.InconsistentEventTypeNEQ)
			}
		default:
			// 记录日志，然后继续
			// 做好监控
			v.l.Error("base -> target 查询 target 失败",
				logger.Int64("id", src.ID()),
				logger.Error(err))
		}
	}
}

func (v *Validator[T]) validateTargetToBase(ctx context.Context) error {
	offset := -v.batchSize
	for {
		offset += v.batchSize
		var ts []T
		err := v.target.WithContext(ctx).Select("id").
			Order("id").Offset(offset).Limit(v.batchSize).
			Find(&ts).Error
		if err == gorm.ErrRecordNotFound || len(ts) == 0 {
			return nil
		}
		if err != nil {
			v.l.Error("target => base 查询 target 失败", logger.Error(err))
			continue
		}
		// 在这里
		var srcTs []T
		ids := slice.Map(ts, func(idx int, t T) int64 {
			return t.ID()
		})
		err = v.base.WithContext(ctx).Select("id").
			Where("id IN ?", ids).Find(&srcTs).Error
		if err == gorm.ErrRecordNotFound || len(srcTs) == 0 {
			// 都代表。base 里面一条对应的数据都没有
			v.notifyBaseMissing(ts)
			continue
		}
		if err != nil {
			v.l.Error("target => base 查询 base 失败", logger.Error(err))
			// 保守起见，我都认为 base 里面没有数据
			// v.notifyBaseMissing(ts)
			continue
		}
		// 找差集，diff 里面的，就是 target 有，但是 base 没有的
		diff := slice.DiffSetFunc(ts, srcTs, func(src, dst T) bool {
			return src.ID() == dst.ID()
		})
		v.notifyBaseMissing(diff)
		// 说明也没了
		if len(ts) < v.batchSize {
			return nil
		}
	}
}

func (v *Validator[T]) notifyBaseMissing(ts []T) {
	for _, val := range ts {
		v.notify(val.ID(), events.InconsistentEventTypeBaseMissing)
	}
}

func (v *Validator[T]) notify(id int64, typ string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := v.producer.ProduceInconsistentEvent(ctx, events.InconsistentEvent{
		ID:        id,
		Type:      typ,
		Direction: v.direction,
	})
	v.l.Error("发送不一致消息失败",
		logger.Error(err),
		logger.String("type", typ),
		logger.Int64("id", id))
}
