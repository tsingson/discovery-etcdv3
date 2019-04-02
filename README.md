# etcdv3 服务发现与负载均衡


## 说明
参考文章来自 [鸟窝-转// gRPC服务发现&负载均衡](https://colobu.com/2017/03/25/grpc-naming-and-load-balance/)

原文出处: [gRPC服务发现&负载均衡](https://segmentfault.com/a/1190000008672912), 作者: [softfn](https://segmentfault.com/u/softfn)。




## movation 动机

为  [goim](https://github.com/Terry-Mao/goim)  的试手项目而改写


## 测试

启动ETCD

```
#https://coreos.com/etcd/docs/latest/op-guide/container.html#docker
	
	export NODE1=192.168.1.21
	
	docker volume create --name etcd-data
	export DATA_DIR="etcd-data"
	
	REGISTRY=quay.io/coreos/etcd
	# available from v3.2.5
	# REGISTRY=gcr.io/etcd-development/etcd
	
	docker run -d\
	  -p 2379:2379 \
	  -p 2380:2380 \
	  --volume=${DATA_DIR}:/etcd-data \
	  --name etcd ${REGISTRY}:latest \
	  /usr/local/bin/etcd \
	  --data-dir=/etcd-data --name node1 \
	  --initial-advertise-peer-urls http://${NODE1}:2380 --listen-peer-urls http://0.0.0.0:2380 \
	  --advertise-client-urls http://${NODE1}:2379 --listen-client-urls http://0.0.0.0:2379 \
	  --initial-cluster node1=http://${NODE1}:2380

#启动测试程序

    # 分别启动服务端
    go run cmd/svr/svr.go - port 50001
    go run cmd/svr/svr.go - port 50002
    go run cmd/svr/svr.go - port 50003
    
#启动客户端
    go run cmd/cli/cli.go
```