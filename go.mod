module github.com/ipweb-group/file-server

go 1.12

require (
	github.com/ethereum/go-ethereum v1.8.7
	github.com/go-redis/redis v6.15.3+incompatible
	github.com/ipfs/go-ipfs v0.4.21
	github.com/ipweb-group/go-sdk v0.9.0
	github.com/kataras/golog v0.0.0-20190624001437-99c81de45f40
	github.com/kataras/iris v11.1.1+incompatible
	gopkg.in/yaml.v2 v2.2.2
)

replace (
	github.com/ethereum/go-ethereum v1.8.7 => github.com/ipweb-group/ipw v1.9.0
	github.com/ipfs/go-ipfs v0.4.21 => github.com/ipweb-group/go-ipws v0.9.3
	github.com/ipweb-group/go-sdk v0.9.0 => /Users/jerry/go/src/github.com/ipweb-group/go-sdk
)