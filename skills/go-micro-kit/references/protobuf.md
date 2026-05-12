# Protobuf and Code Generation

## Workflow

Use a schema-first workflow:

1. Define or update `.proto` files in `proto/`.
2. Generate Go messages, gRPC server/client code, gRPC-Gateway handlers, and OpenAPI.
3. Implement the generated server interface in application handlers.
4. Register the gRPC service with `app.Grpc`.
5. Register the gateway handler with `app.Http`.

## Service Proto Pattern

```proto
syntax = "proto3";

package item.v1;

import "google/api/annotations.proto";
import "protoc-gen-openapiv2/options/annotations.proto";

option go_package = "example/gen/item/v1;itemv1";

option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_swagger) = {
  info: { title: "Item API"; version: "1.0"; }
  host: "localhost:8080"
  schemes: HTTP
};

message CreateItemRequest {
  string name = 1;
}

message CreateItemResponse {
  uint64 id = 1;
  string name = 2;
}

service ItemService {
  rpc CreateItem(CreateItemRequest) returns (CreateItemResponse) {
    option (google.api.http) = {
      post: "/api/items"
      body: "*"
    };
  }
}
```

## Generation Tools

Install protoc plugins:

```bash
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2
go install google.golang.org/protobuf/cmd/protoc-gen-go
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
```

Ensure the install directory, usually `GOBIN` or `GOPATH/bin`, is on `PATH`.

## Generation Command

Adapt paths to the project:

```bash
protoc -I proto ./proto/item/v1/item.proto \
  --go_out=./gen --go_opt=paths=source_relative \
  --go-grpc_out=./gen --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out=./gen --grpc-gateway_opt=paths=source_relative \
  --grpc-gateway_opt=generate_unbound_methods=true \
  --openapiv2_out=./docs --openapiv2_opt=allow_merge=true,merge_file_name=api
```

Expected outputs:

- `gen/**`: generated protobuf, gRPC, and gateway Go files.
- `docs/api.swagger.json`: OpenAPI v2 spec used by Swagger UI.

## Notes

- Commit generated files if the target project already commits generated files.
- Keep `go_package` aligned with the Go module path and generated directory.
- Add gateway annotations for all HTTP-exposed RPCs.
- If using validation annotations, make sure the project vendors or includes the required proto files and has validation middleware configured in gRPC options.

