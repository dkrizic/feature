proto:
        # server part
	protoc -I meta    --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative --go_out=../service/service//meta/v1   --go-grpc_out=../service/service/meta/v1        meta/meta.proto
	protoc -I feature --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative --go_out=../service/service/feature/v1 --go-grpc_out=../service/service/feature/v1 feature/feature.proto
	# client for cli
	protoc -I meta    --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative --go_out=../cli/repository/meta/v1    --go-grpc_out=../cli/repository/meta/v1 meta/meta.proto
	protoc -I feature --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative --go_out=../cli/repository/feature/v1 --go-grpc_out=../cli/repository/feature/v1  feature/feature.proto
	# client for UI
	protoc -I meta    --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative --go_out=../ui/repository/meta//v1   --go-grpc_out=../ui/repository/meta/v1 meta/meta.proto
	protoc -I feature --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative --go_out=../ui/repository/feature/v1 --go-grpc_out=../ui/repository/feature/v1  feature/feature.proto
