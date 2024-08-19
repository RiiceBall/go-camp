package service

import (
	"context"
	"sort"
	"sync"
	"time"

	followv1 "gitee.com/geekbang/basic-go/webook/api/proto/gen/follow/v1"
	"gitee.com/geekbang/basic-go/webook/feed/domain"
	"gitee.com/geekbang/basic-go/webook/feed/repository"
	"github.com/ecodeclub/ekit/slice"
	"golang.org/x/sync/errgroup"
)

type ArticleEventHandler struct {
	repo         repository.FeedEventRepo
	followClient followv1.FollowServiceClient
}

const (
	ArticleEventName = "article_event"
	threshold        = 4
	//threshold        = 32
)

func NewArticleEventHandler(repo repository.FeedEventRepo, client followv1.FollowServiceClient) Handler {
	return &ArticleEventHandler{
		repo:         repo,
		followClient: client,
	}
}

func (h *ArticleEventHandler) FindFeedEvents(ctx context.Context, uid, timestamp, limit int64) ([]domain.FeedEvent, error) {
	// article 这边是要聚合的
	// 可能在 push event，可能在 pull event
	var eg errgroup.Group
	var lock sync.Mutex
	events := make([]domain.FeedEvent, 0, limit*2)
	eg.Go(func() error {
		// 如果一个用户是活跃用户，那么他的数据必然在收件箱里面，就没有必须再去查询了
		if h.isActiveUser(uid) {
			return nil
		}

		// 查询发件箱
		resp, err := h.followClient.GetFollowee(ctx, &followv1.GetFolloweeRequest{Follower: uid, Limit: 10000})
		if err != nil {
			return err
		}
		followeeIDs := slice.Map(resp.FollowRelations, func(idx int, src *followv1.FollowRelation) int64 {
			return src.Followee
		})
		evts, err := h.repo.FindPullEventsWithTyp(ctx, ArticleEventName, followeeIDs, timestamp, limit)
		if err != nil {
			return err
		}
		lock.Lock()
		events = append(events, evts...)
		lock.Unlock()
		return nil
	})

	eg.Go(func() error {
		evts, err := h.repo.FindPushEventsWithTyp(ctx, ArticleEventName, uid, timestamp, limit)
		if err != nil {
			return err
		}
		lock.Lock()
		events = append(events, evts...)
		lock.Unlock()
		return nil
	})

	err := eg.Wait()
	if err != nil {
		return nil, err
	}
	// 你已经查询所有的数据，现在要排序
	sort.Slice(events, func(i, j int) bool {
		return events[i].Ctime.UnixMilli() > events[j].Ctime.UnixMilli()
	})
	return events[:slice.Min[int]([]int{int(limit), len(events)})], nil
}

func (h *ArticleEventHandler) CreateFeedEvent(ctx context.Context, ext domain.ExtendFields) error {
	uid, err := ext.Get("followee").AsInt64()
	if err != nil {
		return err
	}
	// 找到这个人的粉丝数量，判定是拉模型还是推模型
	resp, err := h.followClient.GetFollowStatic(ctx, &followv1.GetFollowStaticRequest{Followee: uid})
	if err != nil {
		return err
	}

	// 大于一个阈值
	if resp.FollowStatic.Followers > threshold {
		// 对活跃的用户进行写扩散
		fresp, err := h.followClient.GetFollower(ctx, &followv1.GetFollowerRequest{Followee: uid})
		if err != nil {
			return err
		}
		events := slice.Map(fresp.FollowRelations, func(idx int, src *followv1.FollowRelation) domain.FeedEvent {
			if h.isActiveUser(src.Follower) {
				return domain.FeedEvent{Uid: src.Follower, Ctime: time.Now(), Type: ArticleEventName, Ext: ext}
			} else {
				return domain.FeedEvent{}
			}
		})
		// 过滤掉空的 FeedEvent
		filteredEvents := slice.FilterDelete(events, func(idx int, event domain.FeedEvent) bool {
			return len(event.Ext) == 0
		})
		err = h.repo.CreatePushEvents(ctx, filteredEvents)
		if err != nil {
			return err
		}

		// 拉模型
		return h.repo.CreatePullEvent(ctx, domain.FeedEvent{Uid: uid,
			Type:  ArticleEventName,
			Ctime: time.Now(),
			Ext:   ext})
	} else {
		// 推模型，也就是写扩散
		// 先查询出来粉丝
		fresp, err := h.followClient.GetFollower(ctx, &followv1.GetFollowerRequest{Followee: uid})
		if err != nil {
			return err
		}
		events := slice.Map(fresp.FollowRelations, func(idx int, src *followv1.FollowRelation) domain.FeedEvent {
			return domain.FeedEvent{Uid: src.Follower, Ctime: time.Now(), Type: ArticleEventName, Ext: ext}
		})
		return h.repo.CreatePushEvents(ctx, events)
	}
}

func (h *ArticleEventHandler) isActiveUser(uid int64) bool {
	// 判断一个用户是否是活跃用户，查看用户在过去一周内的行为，如果同时都满足则判断为活跃用户
	// 1. 是否多次登录过
	// 2. 是否在网站上进行过多次操作（比如是个视频网站的话是否看过视频）
	// 3. 是否浏览过动态页面
	return false
}
