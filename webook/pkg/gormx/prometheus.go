package gormx

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/gorm"
)

type Callbacks struct {
	vector *prometheus.SummaryVec
}

func NewCallbacks(opts prometheus.SummaryOpts) *Callbacks {
	vector := prometheus.NewSummaryVec(opts, []string{"type", "table"})
	prometheus.MustRegister(vector)
	return &Callbacks{
		vector: vector,
	}
}

func (c *Callbacks) Name() string {
	return "prometheus"
}

func (c *Callbacks) Initialize(db *gorm.DB) error {
	// Querys
	err := db.Callback().Query().Before("*").
		Register("prometheus_query_before", c.before("query"))
	if err != nil {
		return err
	}

	err = db.Callback().Query().After("*").
		Register("prometheus_query_after", c.after("query"))
	if err != nil {
		return err
	}

	err = db.Callback().Raw().Before("*").
		Register("prometheus_raw_before", c.before("raw"))
	if err != nil {
		return err
	}

	err = db.Callback().Raw().After("*").
		Register("prometheus_raw_after", c.after("raw"))
	if err != nil {
		return err
	}

	err = db.Callback().Create().Before("*").
		Register("prometheus_create_before", c.before("create"))
	if err != nil {
		return err
	}

	err = db.Callback().Create().After("*").
		Register("prometheus_create_after", c.after("create"))
	if err != nil {
		return err
	}

	err = db.Callback().Update().Before("*").
		Register("prometheus_update_before", c.before("update"))
	if err != nil {
		return err
	}

	err = db.Callback().Update().After("*").
		Register("prometheus_update_after", c.after("update"))
	if err != nil {
		return err
	}

	err = db.Callback().Delete().Before("*").
		Register("prometheus_delete_before", c.before("delete"))
	if err != nil {
		return err
	}

	err = db.Callback().Delete().After("*").
		Register("prometheus_delete_after", c.after("delete"))
	if err != nil {
		return err
	}
	return nil
}

func (c *Callbacks) before(typ string) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		start := time.Now()
		db.Set("start_time", start)
	}
}

func (c *Callbacks) after(typ string) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		val, _ := db.Get("start_time")
		// 如果上面没找到，这边必然断言失败
		start, ok := val.(time.Time)
		if ok {
			duration := time.Since(start).Microseconds()
			c.vector.WithLabelValues(typ, db.Statement.Table).
				Observe(float64(duration))
		}
	}
}
