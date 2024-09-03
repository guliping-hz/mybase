package rpc

import (
	"container/ring"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"sync"
)

type NewClientF func(*grpc.ClientConn) any

type ClientTarget struct {
	target string
	client any
	Conn   *grpc.ClientConn
	//Pool *grpcpool.Pool
}

func (r *ClientTarget) startRPCClient(ctx context.Context, target string, newF NewClientF) {
	r.target = target

	//r.Pool, _ = grpcpool.New(func() (*grpc.ClientConn, error) {
	//	return grpc.DialContext(ctx, target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	//}, 5, 5, 0)

	conn, _ := grpc.DialContext(ctx, target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	r.client, r.Conn = newF(conn), conn
}

type BalanceConnectPool struct {
	lock    sync.Mutex
	Clients *ring.Ring
	Balance int
}

type Clients struct {
	Targets []string
	dict    sync.Map //map[string]*BalanceConnectPool
	ctx     context.Context
}

func NewClients(ctx context.Context, targets []string, except *string, newF NewClientF, poolCnt int) *Clients {
	clients := new(Clients)
	clients.ctx = ctx
	clients.Targets = targets

	for i := range targets {
		if except != nil && *except == targets[i] {
			continue
		}

		//pool := &BalanceConnectPool{Clients: make([]*ClientTarget, poolCnt)}
		pool := &BalanceConnectPool{}
		for j := 0; j < poolCnt; j++ {
			cli := new(ClientTarget)
			cli.startRPCClient(ctx, targets[i], newF)

			//pool.Clients[j] = cli
			if pool.Clients == nil {
				pool.Clients = &ring.Ring{Value: cli}
			} else {
				pool.Clients.Link(&ring.Ring{Value: cli})
			}
		}
		clients.dict.Store(targets[i], pool)
	}
	return clients
}

func (c *Clients) getCli(target string) *ClientTarget {
	cliV, ok := c.dict.Load(target)
	if !ok {
		return nil //RPC定位失败
	}
	pool, ok := cliV.(*BalanceConnectPool)
	if !ok || /*len(pool.Clients)*/ pool.Clients.Len() == 0 {
		return nil //RPC类型存储有问题。
	}

	pool.lock.Lock()
	defer pool.lock.Unlock()

	//if pool.Balance >= len(pool.Clients) {
	//	pool.Balance = 0
	//}
	cli, _ := pool.Clients.Value.(*ClientTarget)
	pool.Clients = pool.Clients.Move(1)
	//pool.Balance++
	return cli
}

// 连接池功能去掉了。。把游戏卡住了。
func (c *Clients) GetCli(target string) (any, *grpc.ClientConn) {
	cli := c.getCli(target)
	if cli == nil {
		return nil, nil
	}
	return cli.client, cli.Conn

	//waitCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	//defer cancel()
	//cc, err := cli.Pool.Get(waitCtx)
	//if err != nil { //RPC连接繁忙
	//	return nil, nil
	//}
	//return survive.NewGameClient(cc), cc
}

func (c *Clients) ShutDown() {
	c.dict.Range(func(key, value any) bool {
		if pool, ok := value.(*BalanceConnectPool); ok {
			pool.Clients.Do(func(v any) {
				cli, _ := v.(*ClientTarget)
				if cli != nil {
					_ = cli.Conn.Close()
				}
			})
			//for i := range pool.Clients {
			//	pool.Clients[i].Conn.Close()
			//}
		}
		return true
	})
	//清理
	c.dict = sync.Map{}
}
