set -e

cd ../..

protoc --go_out=. --go-grpc_out=. internal/fubotorp/*.proto
