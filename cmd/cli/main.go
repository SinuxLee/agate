package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"template/pkg/proto"

	microhttp "github.com/asim/go-micro/plugins/client/http/v3"
	"github.com/asim/go-micro/plugins/registry/consul/v3"
	"github.com/asim/go-micro/plugins/selector/shard/v3"
	"github.com/asim/go-micro/plugins/transport/grpc/v3"
	"github.com/asim/go-micro/plugins/wrapper/breaker/hystrix/v3"
	"github.com/asim/go-micro/plugins/wrapper/trace/opencensus/v3"
	"github.com/asim/go-micro/v3/client"
	"github.com/asim/go-micro/v3/registry"
	"github.com/asim/go-micro/v3/selector"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

var (
	emptyData = struct{}{}

	data  = []byte("giny")
	magic = uint32(data[3]) | (uint32(data[2]) << 8) | (uint32(data[1]) << 16) | (uint32(data[0]) << 24)
)

const (
	consulAddr = "127.0.0.1:8500"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type HelloRsp struct {
	Response
	Data struct {
		Greet string `json:"greet"`
	} `json:"data,omitempty"`
}

func makeToken() {
	var ts, uid uint32
	buff := bytes.NewBuffer(nil)
	buff.Grow(256)

	ts = uint32(time.Now().Unix())
	uid = 1234567890 + rand.Uint32()
	mask := magic ^ ts ^ uid

	binary.Write(buff, binary.BigEndian, uid)
	binary.Write(buff, binary.BigEndian, ts)
	binary.Write(buff, binary.BigEndian, mask)

	fmt.Println(base64.RawURLEncoding.EncodeToString(buff.Bytes()))
}

func parseToken(token string) error {
	data, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return err
	}
	buff := bytes.NewBuffer(data)

	var ts, uid, mask uint32
	binary.Read(buff, binary.BigEndian, &uid)
	binary.Read(buff, binary.BigEndian, &ts)
	binary.Read(buff, binary.BigEndian, &mask)
	xor := magic ^ ts ^ uid

	fmt.Println(ts, uid, xor)
	if xor != mask {
		fmt.Println(mask, xor)
	}

	return nil
}

func benchmark() {
	for {
		now := time.Now()
		rsp, err := http.Get("http://localhost:28086/")
		if err != nil {
			log.Err(err).Msg("request failed")
			continue
		}

		data, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			log.Err(err).Msg("read body failed")
			continue
		}
		log.Info().TimeDiff("cost", time.Now(), now).Msg(string(data))
		rsp.Body.Close()
		time.Sleep(time.Millisecond * 10)
	}
}

func main() {
	webCli()
	rpcCli()

	makeToken()
	parseToken("YyqD9UmWAtJN1e9e")

	benchmark()
}

func webCli() {
	reg := consul.NewRegistry(
		registry.Addrs(consulAddr),
		registry.Timeout(time.Second*10),
	)

	sel := selector.NewSelector(
		selector.Registry(reg),
	)

	cli := microhttp.NewClient(
		client.PoolSize(500),
		client.PoolTTL(time.Minute),
		client.Selector(sel),
		client.Retries(3),
		client.DialTimeout(time.Second*2),
		client.RequestTimeout(time.Second*5),
		client.Wrap(hystrix.NewClientWrapper()),
		client.Wrap(opencensus.NewClientWrapper()),
		client.ContentType(gin.MIMEJSON),
	)

	// 只能调用POST 方法
	rsp := &HelloRsp{}
	req := cli.NewRequest("svrWEB", "/svr/v1/hello/libz", emptyData)
	err := cli.Call(context.TODO(), req, rsp)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(rsp.Data.Greet)
}

func rpcCli() {
	cli := client.NewClient(
		client.Selector(selector.NewSelector(
			selector.Registry(consul.NewRegistry(
				registry.Addrs(consulAddr),
			)),
		)),
		client.Transport(grpc.NewTransport()),
		client.Retries(3),
	)

	// Use the generated client stub
	cl := proto.NewGreeterService("svrRPC", cli) // consul中注册的服务名称

	// Make request
	req := &proto.HelloRequest{
		Name: "libz",
	}
	rsp, err := cl.Hello(context.Background(), req, shard.Strategy(req.Name)) // chash
	if err != nil {
		log.Err(err).Msg("Hello request failed")
		return
	}

	fmt.Println(rsp.Greeting)
}
