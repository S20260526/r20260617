all:
	cd ../../ \
		&& python3 -m grpc_tools.protoc \
			--go_out=. --go-grpc_out=. \
			--python_out=internal/grpcipc/python \
			--grpc_python_out=internal/grpcipc/python \
			--proto_path=internal/grpcipc/proto \
			ipc.proto
