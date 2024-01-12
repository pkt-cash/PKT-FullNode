module github.com/pkt-cash/PKT-FullNode

go 1.18

replace git.schwanenlied.me/yawning/bsaes.git => github.com/Yawning/bsaes v0.0.0-20180720073208-c0276d75487e

replace github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.5

replace google.golang.org/grpc => google.golang.org/grpc v1.29.1

require (
	github.com/aead/chacha20 v0.0.0-20180709150244-8b13a72661da
	github.com/aead/siphash v1.0.1
	github.com/arl/statsviz v0.2.2-0.20201115121518-5ea9f0cf1bd1
	github.com/btcsuite/winsvc v1.0.0
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc
	github.com/dchest/blake2b v1.0.0
	github.com/golang/snappy v0.0.2
	github.com/gorilla/websocket v1.4.3-0.20200912193213-c3dd95aea977
	github.com/jessevdk/go-flags v1.4.1-0.20200711081900-c17162fe8fd7
	github.com/json-iterator/go v1.1.11-0.20200806011408-6821bec9fa5c
	github.com/kkdai/bstream v1.0.0
	github.com/onsi/ginkgo v1.14.3-0.20201013214636-dfe369837f25
	github.com/onsi/gomega v1.10.3
	github.com/sethgrid/pester v1.1.1-0.20200617174401-d2ad9ec9a8b6
	github.com/stretchr/testify v1.6.1
	golang.org/x/crypto v0.13.0
)

require (
	github.com/cjdelisle/Electorium_go v0.0.0-20231004103752-531e568a992b // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/fsnotify/fsnotify v1.4.10-0.20200417215612-7f4cf4dd2b52 // indirect
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/nxadm/tail v1.4.6-0.20201001195649-edf6bc2dfc36 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/exp v0.0.0-20230905200255-921286631fa9 // indirect
	golang.org/x/net v0.10.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/protobuf v1.24.0 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
)
