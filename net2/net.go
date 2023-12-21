package net2

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrTimeout = errors.New("time out")
	ErrParam   = errors.New("param error")
	ErrBuffer  = errors.New("buffer error")
	ErrClose   = errors.New("closed by the peer")
	ErrOOM     = errors.New("oom")
)

/*
*
对外宣称对象
*/
type Conn interface {
	//获取最近的错误
	error

	//安全关闭连接
	//@param byLocalNotRemote 是否是本地主动本地断开；true:对服务器而言就是服务器把客户端断开，对客户端而言就是客户端主动断服务器
	SafeClose(byLocalNotRemote bool)

	// 封装发送buffer
	Send(buf []byte) bool

	//是否是被对方关闭了连接
	IsClosedByPeer() bool

	//获取最近的调用堆栈，如果有的话
	Stack() []byte

	//本地地址
	LocalAddr() net.Addr

	//对方地址
	RemoteAddr() net.Addr

	//当前服务器中的唯一ID
	SessionId() uint64

	//唯一ID，多个服务器的话可能 SessionId会一样，但是UnionId肯定不一样
	UnionId() string
}

// net2内部封装接口，不对外开放
type iSocket interface {
	Close() error

	sendEx(buffer []byte)

	recvEx() ([]byte, error)
}

/*
*
Socket 事件分发
*/
type OnSocket interface {
	/**
	连接上服务器回调,或者服务器accept某个客户端连接
	*/
	OnConnect(conn Conn)
	/**
	只要我们曾经连接上服务器过，OnClose必定会回调。代表一个当前的socket已经关闭
	@param conn 连接
	@param byLocalNotRemote 是否是本地主动本地断开；true:对服务器而言就是服务器把客户端断开，对客户端而言就是客户端主动断服务器
	*/
	OnClose(conn Conn, byLocalNotRemote bool)
	/**
	连接超时,写入超时,读取超时回调
	*/
	OnTimeout(conn Conn)
	/**
	网络错误回调，之后直接close
	*/
	OnNetErr(conn Conn)
	/**
	接受到信息
	@return 返回true表示可以继续热恋，false表示要分手了。
	*/
	OnRecvMsg(conn Conn, buf []byte) bool
}

/*
*
Socket 数据解析
*/
type DataDecodeBase interface {
	//包解析 =》 包长+包内容  =》整个包的长度 = 包长+包内容长度
	//返回包长的长度
	GetPackageHeadLen() int
	//返回包内容的长度
	GetPackageLen([]byte) int
}

type ClientBase struct {
	context *Context
	Status  Status

	chanSendDB  chan []byte
	isConnected int32 //-1连接已关闭，0未连接，1已连接 大于1表示有发送数据占用着，暂时不能关闭
	//sendUse     sync.Map
	SessionIdU uint64
	UnionIdStr string
}

func (c *ClientBase) Error() string {
	return c.Status.Error()
}

func (c *ClientBase) Stack() []byte {
	return c.Status.Stack()
}

func (c *ClientBase) SessionId() uint64 {
	return c.SessionIdU
}

func (c *ClientBase) UnionId() string {
	return c.UnionIdStr
}

func (c *ClientBase) Init(ddb DataDecodeBase, ttl, RTtl time.Duration, onSocket OnSocket, con Conn, socket iSocket) {
	c.context = &Context{}
	c.context.readDB = &bytes.Buffer{}
	c.context.dataDecoder = ddb
	if c.context.dataDecoder == nil {
		c.context.dataDecoder = new(DataDecodeBinaryBigEnd)
	}
	c.context.ttl = ttl
	c.context.rTtl = RTtl

	c.context.Con = con
	c.context.socket = socket
	c.context.OnSocket = onSocket
}

func (c *ClientBase) SafeClose(byLocalNotRemote bool) {
	//c.mutexConnect.Lock() //=》死锁 由于折返锁的原因，这里被死锁了  0_0!!
	//defer c.mutexConnect.Unlock()
	//mybase.D("SafeClose")

	for {
		swapped := atomic.CompareAndSwapInt32(&c.isConnected, 1, -1)
		if swapped { //如果允许关闭
			break
		} else {
			if atomic.LoadInt32(&c.isConnected) == -1 {
				return
			}

			//fmt.Printf("session=%d SafeClose\n", c.SessionId())
			go func() { //防止外面死锁。。。
				time.Sleep(time.Second) //等待1 ms
				c.SafeClose(byLocalNotRemote)
			}()
			return
		}
	}

	c.Status.ChangeStatus(StatusShutdown, nil)
	err := c.context.socket.Close()
	if err != nil {
		//mybase.E("Close error=%v", err.Error())
	}

	close(c.context.chanStop)
	close(c.chanSendDB)

	//mybase.D("SafeClose closed %d", atomic.LoadInt32(&c.isConnected))

	c.safeSendOnClose(byLocalNotRemote) //如果需要回调，我们就回调一下。
	c.context.readDB.Reset()            //清空已读的buffer
}

func (c *ClientBase) Send(buf []byte) bool {
	//c.mutexConnect.Lock()
	//defer c.mutexConnect.Unlock()
	//mybase.D("Send buf 1")

	defer atomic.AddInt32(&c.isConnected, -1)    //安全的发送数据
	if atomic.AddInt32(&c.isConnected, 1) <= 1 { //如果当前是连接状态，这个数一定是大于1的。
		return false
	}

	//mybase.D("Send buf 2")

	c.chanSendDB <- buf
	return true
}

//实验二
//@注意：下面的两个实现最终失败了，，可能会发生 panic： send on closed

//实验二： SafeClose实现
//func (c *ClientBase) SafeClose(bySelf bool) {
//	//mybase.LOG.Tracef("SafeClose 1 %d con=%d", c.SessionId(), c.isConnected)
//	if c.isConnected == 0 {
//		return
//	}
//	c.isConnected = 0
//
//	c.context.once.Do(func() { //这里会死锁。。 客户端断开->对等断开->再回到客户端断开,导致死锁
//		//mybase.I("SafeClose 2 %d", c.SessionId())
//		c.Status.ChangeStatus(StatusShutdown, nil)
//		err := c.context.socket.Close()
//		if err != nil {
//			mybase.E("Close error=%v", err.Error())
//		}
//
//		close(c.context.chanStop)
//		c.safeSendOnClose(bySelf) //如果需要回调，我们就回调一下。
//		c.context.readDB.Reset()  //清空已读的buffer
//	})
//	//mybase.I("SafeClose 3 %d", c.SessionId())
//}

//实验二： Send实现
//func (c *ClientBase) Send(buf []byte) bool {
//	select { //即使是这样，在极端情况下也会发生panic send on closed
//	case <-c.context.Done():
//		return false
//	default:
//		c.chanSendDB <- buf
//		return true
//	}
//}

//实验一： Send实现
//func (c *ClientBase) Send(buf []byte) bool {
//	select {
//	case <-c.context.Done():
//		return false
//	default:
//		//直接发送的话，在websocket中会崩溃 concurrent write to websocket connection
//		//ws 要求必须顺序写入buf
//		c.context.socket.SendEx(buf)
//		return true
//	}
//}

func (c *ClientBase) IsClosedByPeer() bool {
	return c.Status.err == ErrClose
}

func (c *ClientBase) CloseWithErr(err error, stack []byte, check bool) {
	ret := c.Status.ChangeStatusAll(StatusError, err, stack)
	if (check && ret) || !check {
		c.context.OnSocket.OnNetErr(c.context.Con)
		//把发生错误的socket及时关闭
		c.context.Con.SafeClose(false)
	}
}

func (c *ClientBase) CloseTimeout() {
	c.Status.ChangeStatus(StatusTimeout, nil)
	c.context.OnSocket.OnTimeout(c.context.Con)
	//把发生错误的socket及时关闭
	c.context.Con.SafeClose(false)
}

func (c *ClientBase) process() error {
	defer func() {
		c.context.readDB = bytes.NewBuffer(c.context.readDB.Bytes()) //舍去已经读取的buffer，保留尚未读取的buffer
	}()

	lenHead := c.context.dataDecoder.GetPackageHeadLen()

	for {
		readLen := c.context.readDB.Len()
		if readLen <= 0 || readLen < lenHead { //不足包长
			return nil
		}

		//if os.Getenv("name") == "robot" {
		//	//mybase.D("recv buf=%v", c.context.readDB.Bytes())
		//}

		lenPackage := c.context.dataDecoder.GetPackageLen(c.context.readDB.Bytes())
		if lenPackage == 0 { //异常包
			//mybase.W("lenPackage=0 readLen=%d,lenHead=%d", readLen, lenHead)
			c.context.readDB.Reset()
			return ErrBuffer
		}

		lenFull := lenHead + lenPackage
		if lenFull > c.context.readDB.Len() { //不足一个包
			//mybase.D("lenFull=%d,lenHead=%d need more %d", lenFull, lenHead, c.context.readDB.Len())
			return nil
		}

		packageBuf := make([]byte, lenFull)
		_, _ = c.context.readDB.Read(packageBuf)
		//mybase.D("read copy buf len=%d", lenFull)
		ok := c.context.OnSocket.OnRecvMsg(c.context.Con, packageBuf)
		if !ok {
			c.context.readDB.Reset()
			return io.EOF
		}
	}
}

func (c *ClientBase) safeSendOnClose(byLocalNotRemote bool) {
	defer func() {
		recover()
	}()

	//通知关闭，这里增加一个recover，防止崩溃
	c.context.OnSocket.OnClose(c.context.Con, byLocalNotRemote) //如果需要回调，我们就回调一下。
}

func (c *ClientBase) Reactor() {
	c.isConnected = 1
	c.Status.Reset() //先清空之前的状态信息
	c.Status.ChangeStatus(StatusNormal, nil)

	c.chanSendDB = make(chan []byte)
	c.context.chanStop = make(chan struct{})
	c.context.once = sync.Once{}

	go c.sendRoutine() //发送协程：按顺序统一发送buff
	go c.recvRoutine() //接收协程:按顺序统一接收buff
}

func (c *ClientBase) sendRoutine() {
	for {
		select {
		case buf := <-c.chanSendDB:
			//fmt.Printf("session=%d sendRoutine 1\n", c.SessionId())
			c.context.socket.sendEx(buf)
			//fmt.Printf("session=%d sendRoutine 2\n", c.SessionId())
		case <-c.context.Done():
			//close(c.chanSendDB) //这里是配合实验二
			//fmt.Printf("session=%d sendRoutine end\n", c.SessionId())
			return
		}
	}
}

func (c *ClientBase) recvRoutine() {
	defer func() {
		p := recover()
		if err, ok := p.(error); ok {
			log.Printf("CloseWithErr err=%s\n", err)
			//异常报错导致的断开连接。。。
			c.CloseWithErr(err, debug.Stack(), true)
		}
	}()

	c.context.OnSocket.OnConnect(c.context.Con)
	for {
		buffer, err := c.context.socket.recvEx()
		if err != nil {
			err1, ok := err.(*net.OpError)
			errStr := err.Error()
			if err == io.EOF || (runtime.GOOS == "windows" &&
				strings.Contains(errStr, "An existing connection was forcibly closed by the remote host") ||
				strings.Contains(errStr, "An established connection was aborted by the software in your host machine")) ||
				strings.Contains(errStr, "connection reset by peer") {
				/*
					1.io.EOF
						正常关闭.指客户端读完服务器发送的数据然后close

					2.
					connection reset by peer(linux)
					An existing connection was forcibly closed by the remote host(windows)
						表示客户端 【没有读取/读取部分】就close

					3.An established connection was aborted by the software in your host machine(windows)
						表示服务器发送数据，客户端已经close,这个经过测试只有在windows上才会出现。linux试了很多遍都是返回io.EOF错误
						解决办法就是客户端发送数据的时候需要wait一下，然后再close，这样close的结果就是2了
				*/
				c.CloseWithErr(ErrClose, nil, true)
			} else if ok && err1 != nil && err1.Timeout() {
				c.CloseTimeout()
			} else {
				//检查是否已经更改了状态，如果已经更改表示是客户端主动close
				c.CloseWithErr(err, nil, true)
			}
			return
		}

		_, err = c.context.readDB.Write(buffer)
		if err != nil {
			c.CloseWithErr(ErrOOM, nil, true) //无法把buffer全部塞进去，多半是没有内存了。
			return
		}

		err = c.process()
		if err != nil {
			c.CloseWithErr(err, nil, true)
			return
		}
	}
}
