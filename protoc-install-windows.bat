go get github.com/gogo/protobuf/proto
go get github.com/gogo/protobuf/protoc-gen-gogo
go get github.com/gogo/protobuf/protoc-gen-gofast
go get github.com/gogo/protobuf/protoc-gen-gogofast
go get github.com/gogo/protobuf/gogoproto

cp ../github.com/gogo/protobuf/gogoproto/gogo.proto . 
cp ../github.com/gogo/protobuf/protobuf/google/protobuf/descriptor.proto . 

sed -i -e 's:google/protobuf/descriptor.proto:descriptor.proto:g' gogo.proto
protoc --proto_path=.;common/ --gogofast_out=. common/model.proto

rm gogo.proto
rm descriptor.proto

sed -i -e 's:import _ "."::g' common/model.pb.go common/modelpb_test.go
