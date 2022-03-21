package chat

import (
	"Open_IM/pkg/common/mq/nsq"
	"net"
	"strconv"
	"strings"

	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/common/mq"
	"Open_IM/pkg/common/mq/kafka"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbChat "Open_IM/pkg/proto/chat"
	"Open_IM/pkg/utils"

	"google.golang.org/grpc"
)

type rpcChat struct {
	rpcPort         int
	rpcRegisterName string
	etcdSchema      string
	etcdAddr        []string
	producer        mq.Producer
}

func NewRpcChatServer(port int) *rpcChat {
	log.NewPrivateLog("msg")
	rc := rpcChat{
		rpcPort:         port,
		rpcRegisterName: config.Config.RpcRegisterName.OpenImOfflineMessageName,
		etcdSchema:      config.Config.Etcd.EtcdSchema,
		etcdAddr:        config.Config.Etcd.EtcdAddr,
	}
	cfg := config.Config.MQ.Ws2mschat
	switch cfg.Type {
	case "kafka":
		rc.producer = kafka.NewKafkaProducer(cfg.Addr, cfg.Topic)
	case "nsq":
		p, err := nsq.NewNsqProducer(cfg.Addr[0], cfg.Topic)
		if err != nil {
			panic(err)
		}
		rc.producer = p
	}
	return &rc
}

func (rpc *rpcChat) Run() {
	log.Info("", "", "rpc get_token init...")

	address := utils.ServerIP + ":" + strconv.Itoa(rpc.rpcPort)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Error("", "", "listen network failed, err = %s, address = %s", err.Error(), address)
		return
	}
	log.Info("", "", "listen network success, address = %s", address)

	//grpc server
	srv := grpc.NewServer()
	defer srv.GracefulStop()

	//service registers with etcd

	pbChat.RegisterChatServer(srv, rpc)
	err = getcdv3.RegisterEtcd(rpc.etcdSchema, strings.Join(rpc.etcdAddr, ","), utils.ServerIP, rpc.rpcPort, rpc.rpcRegisterName, 10)
	if err != nil {
		log.Error("", "", "register rpc get_token to etcd failed, err = %s", err.Error())
		return
	}

	err = srv.Serve(listener)
	if err != nil {
		log.Info("", "", "rpc get_token fail, err = %s", err.Error())
		return
	}
	log.Info("", "", "rpc get_token init success")
}
