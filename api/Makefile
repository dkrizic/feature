proto:
	protoc -I meta    --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative --go_out=../service/service/meta/metav1       --go-grpc_out=../service/service/meta/metav1        meta/meta.proto
	protoc -I feature --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative --go_out=../service/service/feature/featurev1 --go-grpc_out=../service/service/feature/featurev1 feature/feature.proto
