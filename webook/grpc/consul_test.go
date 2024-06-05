package grpc

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
)

const consulScheme = "consul"

var errUnknownScheme = errors.New("unknown scheme. Only 'consul' is applicable")

// 用于服务发现
type consulResolver struct {
	client *api.Client
	cc     resolver.ClientConn
}

func newConsulResolver(client *api.Client, cc resolver.ClientConn) *consulResolver {
	return &consulResolver{
		client: client,
		cc:     cc,
	}
}

func (r *consulResolver) ResolveNow(opts resolver.ResolveNowOptions) {}

func (r *consulResolver) Close() {}

func (r *consulResolver) start(serviceName string) {
	go func() {
		// 查询服务信息，需要和服务端的服务名一致
		entries, _, err := r.client.Health().Service(serviceName, "", true, nil)
		if err != nil {
			log.Fatalf("Failed to fetch service entries: %v", err)
		}

		addresses := make([]resolver.Address, len(entries))

		for i, entry := range entries {
			// 保存服务地址
			addresses[i] = resolver.Address{Addr: fmt.Sprintf("%s:%d", entry.Service.Address, entry.Service.Port)}
		}

		r.cc.UpdateState(resolver.State{Addresses: addresses})
	}()
}

// 用于构建 resolver
type consulBuilder struct {
	client *api.Client
}

func newConsulBuilder(client *api.Client) *consulBuilder {
	return &consulBuilder{client: client}
}

// 解析 url
func parseTarget(target resolver.Target) (string, error) {
	u, err := url.Parse(fmt.Sprintf("%s://%s", consulScheme, target.Endpoint()))
	if err != nil {
		return "", fmt.Errorf("parse: %w", err)
	}
	return strings.Trim(u.Host, "/"), nil
}

// 启动服务发现
func (b *consulBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	if target.URL.Scheme != consulScheme {
		return nil, errUnknownScheme
	}

	serviceName, err := parseTarget(target)
	if err != nil {
		return nil, fmt.Errorf("parse target: %w", err)
	}

	cr := newConsulResolver(b.client, cc)
	cr.start(serviceName)
	return cr, nil
}

func (b *consulBuilder) Scheme() string {
	return consulScheme
}

type ConsulTestSuite struct {
	suite.Suite
}

func (s *ConsulTestSuite) TestClient() {
	t := s.T()
	// 创建 Consul 客户端
	config := api.DefaultConfig()
	client, err := api.NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create Consul client: %v", err)
	}

	builder := newConsulBuilder(client)

	cc, err := grpc.Dial("consul://127.0.0.1:8500/consul-test",
		grpc.WithResolvers(builder),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	uc := NewUserServiceClient(cc)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := uc.GetByID(ctx, &GetByIDRequest{Id: 123})
	require.NoError(t, err)
	t.Log(resp.User)
}

func (s *ConsulTestSuite) TestServer() {
	t := s.T()
	addr := "localhost:8090"
	l, err := net.Listen("tcp", addr)
	require.NoError(t, err)

	server := grpc.NewServer()
	RegisterUserServiceServer(server, &Server{})
	go func() {
		err := server.Serve(l)
		require.NoError(t, err)
	}()

	config := api.DefaultConfig()
	client, err := api.NewClient(config)
	require.NoError(t, err)

	registration := &api.AgentServiceRegistration{
		ID:      "consul-test",
		Name:    "consul-test",
		Address: "localhost",
		Port:    8090,
	}
	err = client.Agent().ServiceRegister(registration)
	require.NoError(t, err)

	time.Sleep(time.Second * 10)

	defer client.Agent().ServiceDeregister("consul-test")
	defer server.Stop()
}

func TestConsul(t *testing.T) {
	suite.Run(t, new(ConsulTestSuite))
}
