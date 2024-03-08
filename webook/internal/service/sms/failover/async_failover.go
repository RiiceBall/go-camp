package failover

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"
	"webook/internal/domain"
	"webook/internal/repository"
	"webook/internal/service/sms"
)

type AsyncFailOverSMSService struct {
	svc     sms.Service
	smsRepo repository.SmsRepository

	// 错误阈值
	errorThreshold int32
	// 错误计数器
	errorCount int32
	// 错误计数器的时间窗口
	windowDuration time.Duration
	// 上次重置错误计数器的时间
	lastResetTime time.Time

	// 异步发送已启动
	asyncSendStarted bool
}

func NewAsyncFailOverSMSService(svc sms.Service, smsRepo repository.SmsRepository,
	errorThreshold int32, windowDuration time.Duration) *AsyncFailOverSMSService {
	return &AsyncFailOverSMSService{
		svc:            svc,
		smsRepo:        smsRepo,
		errorThreshold: errorThreshold,
		errorCount:     0,
		windowDuration: windowDuration,
		lastResetTime:  time.Now(),
	}
}

/*
 * 设计思路：
 * 查看一段时间内的错误率，如果过高，就将请求添加到数据库
 *
 * 适用场景：
 * 短时间内收到大量的发送请求，如果服务商出现问题，就会导致大量的发送失败
 *
 * 优点：
 * 1. 实现简单：需要在一定的时间内判断错误率是否过高
 * 2. 灵活性高：可以根据实际情况调整错误阈值和时间窗口
 *
 * 缺点：
 * 1. 没有分析错误的原因，只是简单的判断错误率，可能会有误判
 *
 * 优化方案：
 * 1. 可以根据错误的类型进行不同的处理
 * 2. 增加更多的监控指标，比如平均请求的响应时间
 */
func (a *AsyncFailOverSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	now := time.Now()

	// 检查时间窗口是否应该重置
	if now.Sub(a.lastResetTime) > a.windowDuration {
		atomic.StoreInt32(&a.errorCount, 0) // 重置错误计数器
		a.lastResetTime = now               // 更新重置时间
	}

	err := a.svc.Send(ctx, tplId, args, numbers...)
	if err != nil {
		// 记录错误
		atomic.AddInt32(&a.errorCount, 1)

		// 判断错误是否连续发生
		if atomic.LoadInt32(&a.errorCount) > a.errorThreshold {
			// 错误率快速上升，可以认为服务可能崩溃，就将请求添加到数据库
			err := a.smsRepo.Create(ctx, domain.Sms{
				TplId:     tplId,
				Args:      args,
				Numbers:   numbers,
				RetryLeft: 3, // 重试次数设置为 3 次
			})
			if err == nil {
				// 如果异步发送没有启动，就启动异步发送
				if !a.asyncSendStarted {
					a.asyncSendStarted = true
					go a.startAsyncSend(ctx)
				}
			}
			return err
		}
	}
	return err
}

func (a *AsyncFailOverSMSService) startAsyncSend(ctx context.Context) {
	// 设置一个定时器，每隔一段时间检查数据库中是否有未发送的短信
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	// 用于通知异步发送已经结束
	done := make(chan struct{})

	var mutex sync.Mutex
	for {
		select {
		case <-ticker.C:
			go func() {
				// 保证只有一个 goroutine 在处理数据库中的数据
				mutex.Lock()
				defer mutex.Unlock()

				sms, err := a.smsRepo.FindFirstSms(ctx)
				if err == repository.ErrSmsNotFound {
					// 数据库中没有数据，中断异步
					close(done)
					return
				}
				if err != nil {
					// 其他错误，则单纯的返回
					return
				}
				err = a.svc.Send(ctx, sms.TplId, sms.Args, sms.Numbers...)
				if err == nil || sms.RetryLeft <= 1 {
					// 发送成功或重试次数耗尽，删除记录
					err = a.smsRepo.DeleteById(ctx, sms.Id)
					if err != nil {
						// 删除失败，记录日志
						log.Println(err)
					}
					return
				}
				if err != nil {
					// 更新重试次数
					err = a.smsRepo.UpdateRetryLeft(ctx, sms.Id, sms.RetryLeft-1)
					if err != nil {
						// 更新失败，记录日志
						log.Println(err)
					}
				}
			}()
		case <-done:
			a.asyncSendStarted = false
			return
		}

	}
}
