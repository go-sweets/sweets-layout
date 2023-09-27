package service

import (
	"github.com/go-sweets/sweets-layout/internal/boundedcontexts/hello"
	"github.com/google/wire"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(hello.InjectHelloRepository, hello.InjectHelloGrpcHandler, NewHelloServer)
