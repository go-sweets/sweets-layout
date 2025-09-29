package middleware

import (
	"fmt"

	hello_pb "github.com/go-sweets/sweets-layout/api/gen/go/proto"
	hello_kitex "github.com/go-sweets/sweets-layout/api/gen/kitex/api/hello"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s': %s", e.Field, e.Message)
}

// ValidateKitexHelloReq validates a Kitex HelloReq by converting to protobuf and validating
func ValidateKitexHelloReq(req *hello_kitex.HelloReq) error {
	if req == nil {
		return ValidationError{Field: "request", Message: "request cannot be nil"}
	}

	// Convert Kitex type to standard protobuf type
	pbReq := &hello_pb.HelloReq{
		Id: req.Id,
	}

	// Validate using the generated validation method
	if err := pbReq.Validate(); err != nil {
		return ValidationError{Field: "id", Message: err.Error()}
	}

	return nil
}

// ValidateKitexHelloResp validates a Kitex HelloResp by converting to protobuf and validating
func ValidateKitexHelloResp(resp *hello_kitex.HelloResp) error {
	if resp == nil {
		return ValidationError{Field: "response", Message: "response cannot be nil"}
	}

	// Convert Kitex type to standard protobuf type
	pbResp := &hello_pb.HelloResp{
		Id:      resp.Id,
		Message: resp.Message,
	}

	// Validate using the generated validation method
	if err := pbResp.Validate(); err != nil {
		return ValidationError{Field: "response", Message: err.Error()}
	}

	return nil
}

// ValidateHTTPHelloRequest validates HTTP request data
func ValidateHTTPHelloRequest(req interface{}) error {
	type HelloRequest struct {
		Id int64 `json:"id"`
	}

	// Type assert to HelloRequest
	helloReq, ok := req.(*HelloRequest)
	if !ok {
		return ValidationError{Field: "request", Message: "invalid request type"}
	}

	if helloReq == nil {
		return ValidationError{Field: "request", Message: "request cannot be nil"}
	}

	// Convert to protobuf type for validation
	pbReq := &hello_pb.HelloReq{
		Id: helloReq.Id,
	}

	// Validate using the generated validation method
	if err := pbReq.Validate(); err != nil {
		return ValidationError{Field: "id", Message: err.Error()}
	}

	return nil
}
