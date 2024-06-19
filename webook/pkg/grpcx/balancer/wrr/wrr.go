package wrr

import (
	"sync"
	"time"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const Name = "custom_weighted_round_robin"

func newBuilder() balancer.Builder {
	return base.NewBalancerBuilder(Name, &PickerBuilder{}, base.Config{HealthCheck: true})
}

func init() {
	balancer.Register(newBuilder())
}

type PickerBuilder struct {
}

func (p *PickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]*weightConn, 0, len(info.ReadySCs))
	for sc, sci := range info.ReadySCs {
		md, _ := sci.Address.Metadata.(map[string]any)
		weightVal, _ := md["weight"]
		weight, _ := weightVal.(float64)
		conns = append(conns, &weightConn{
			SubConn:         sc,
			weight:          int(weight),
			currentWeight:   int(weight),
			efficientWeight: int(weight),
			available:       true,
		})
	}
	return &Picker{
		conns: conns,
	}
}

type Picker struct {
	conns []*weightConn
	lock  sync.Mutex
}

func (p *Picker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if len(p.conns) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	// 总权重
	var total int
	var maxCC *weightConn
	for _, c := range p.conns {
		if !c.available {
			continue
		}
		c.mutex.Lock()
		total = total + c.weight
		c.currentWeight = c.currentWeight + c.efficientWeight
		if maxCC == nil || maxCC.currentWeight < c.currentWeight {
			maxCC = c
		}
		c.mutex.Unlock()
	}
	maxCC.mutex.Lock()
	maxCC.currentWeight = maxCC.currentWeight - total
	maxCC.mutex.Unlock()

	return balancer.PickResult{
		SubConn: maxCC.SubConn,
		Done: func(info balancer.DoneInfo) {
			maxCC.mutex.Lock()
			defer maxCC.mutex.Unlock()
			if info.Err == nil {
				// 如果成功了，并且当前有效权重小于总权重，就额外增加总权重的 10%
				if maxCC.efficientWeight < total {
					maxCC.efficientWeight += (total / 10)
				}
				return
			}
			code := status.Code(info.Err)
			switch code {
			case codes.Unavailable:
				// 触发熔断
				maxCC.available = false
				go func() {
					// 5 秒后恢复
					time.Sleep(5 * time.Second)
					maxCC.mutex.Lock()
					maxCC.available = true
					// 降低有效权重避免恢复后立刻被太多请求搞的再次熔断
					maxCC.efficientWeight = maxCC.weight / 10
					maxCC.mutex.Unlock()
				}()
			default:
				// 其他错误默认为限流
				// 降低有效权重
				maxCC.efficientWeight -= (total / 10)
				// 有效权重不能小于 1
				if maxCC.efficientWeight < 1 {
					maxCC.efficientWeight = 1
				}
			}
		},
	}, nil
}

type weightConn struct {
	balancer.SubConn

	mutex sync.Mutex

	weight          int
	currentWeight   int
	efficientWeight int

	available bool
}
