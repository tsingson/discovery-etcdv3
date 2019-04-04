
#!/bin/sh
# proto.sh

# descriptor.proto
gopath=$HOME/git/linksmart//src
# gogo.proto
gogopath=${gopath}/github.com/gogo/protobuf/protobuf
# Mcommon.proto等号后面的值，用于把test.proto中import "common.proto"生成为 import "protocol/common"
protoc --proto_path=$gopath:$gogopath:./ --gogo_out=Mcommon.proto=protocol/common:. ./ test.proto


