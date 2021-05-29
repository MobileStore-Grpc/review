import-proto:
	mkdir -p proto/import && \
	go mod download && \
	go list -f "proto/import/{{ .Path }}/proto" -m all \
	| grep proto/import/github.com/MobileStore-Grpc/product/proto | xargs -L1 dirname | sort | uniq | xargs mkdir -p && \
	go list -f "{{ .Dir }}/proto proto/import/{{ .Path }}/proto" -m all \
  	| grep proto/import/github.com/MobileStore-Grpc/product/proto | xargs -L1 -- ln -s

delete-proto-import: 
	find  proto/import -type l -delete && \
	find proto/import -type d -empty -delete

gen:
	protoc -I=proto/ -I=proto/import/github.com/MobileStore-Grpc/product/proto/ \
	--go_out=. --go_opt=module=github.com/MobileStore-Grpc/review \
	--go-grpc_out=. --go-grpc_opt=module=github.com/MobileStore-Grpc/review \
	--grpc-gateway_out=. --grpc-gateway_opt=module=github.com/MobileStore-Grpc/review \
	--openapiv2_out=swagger \
	proto/*.proto

clean:
	rm -r pb/*.go swagger/*

server:
	go run cmd/server/main.go --port 8081

rest:
	go run cmd/server/main.go --port 8082 --type rest --endpoint 0.0.0.0:8081

client:
	go run cmd/client/main.go --address 0.0.0.0:8081


build-image:
	docker build -t mobilestore-review:v1.0.0 .

run:
	docker run -d --name review -p 9010:8010 mobilestore-review:v1.0.0


	# go run cmd/server/main.go --port 8081 --mobileserver 0.0.0.0:8080
