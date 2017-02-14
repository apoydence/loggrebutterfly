package loggrebutterfly

//go:generate bash -c "protoc *.proto -I=. -I=../loggregator/v2/ --go_out=plugins=grpc:."
