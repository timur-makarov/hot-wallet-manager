package proto

//go:generate sh -c "rm -rf gen && mkdir -p gen/common gen/ethereum"
//go:generate sh -c "GOBIN=$(pwd)/../.bin go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"
//go:generate sh -c "PATH=$(pwd)/../.bin:$PATH protoc --proto_path=spec --go_out=paths=source_relative:gen/common --go_opt=MCommon.proto=github.com/timur-makarov/hot-wallet-manager/proto/gen/common --go_opt=MEthereum.proto=github.com/timur-makarov/hot-wallet-manager/proto/gen/ethereum spec/Common.proto"
//go:generate sh -c "PATH=$(pwd)/../.bin:$PATH protoc --proto_path=spec --go_out=paths=source_relative:gen/ethereum --go_opt=MCommon.proto=github.com/timur-makarov/hot-wallet-manager/proto/gen/common --go_opt=MEthereum.proto=github.com/timur-makarov/hot-wallet-manager/proto/gen/ethereum spec/Ethereum.proto"
