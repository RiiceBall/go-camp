package repository

import (
	"context"
	"gitee.com/geekbang/basic-go/webook/internal/domain"
)

type HistoryRecordRepository interface {
	AddRecord(ctx context.Context, record domain.HistoryRecord) error
}
