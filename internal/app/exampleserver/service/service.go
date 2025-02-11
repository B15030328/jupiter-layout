package service

import (
	"context"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	helloworldv1 "github.com/douyu/jupiter-layout/api/helloworld/v1"
	"github.com/douyu/jupiter-layout/internal/pkg/grpc"
	"github.com/douyu/jupiter-layout/internal/pkg/mysql"
	"github.com/douyu/jupiter-layout/internal/pkg/redis"
	"github.com/douyu/jupiter-layout/internal/pkg/rocketmq"

	// "github.com/douyu/jupiter-layout/internal/pkg/redis"
	"github.com/douyu/jupiter-layout/internal/pkg/resty"
	"github.com/douyu/jupiter/pkg/util/xerror"
	"github.com/douyu/jupiter/pkg/xlog"
	"github.com/google/wire"
	"go.uber.org/zap"
)

var ProviderSet = wire.NewSet(
	NewHelloWorldService,
	wire.Struct(new(Options), "*"),
	redis.ProviderSet,
	mysql.ProviderSet,
	grpc.ProviderSet,
	resty.ProviderSet,
	rocketmq.ProviderSet,
)

// Options wireservice
type Options struct {
	ExampleGrpc     grpc.ExampleInterface
	ExampleMysql    mysql.ExampleInterface
	ExampleRedis    redis.ExampleInterface
	ExampleResty    resty.ExampleInterface
	ExampleRocketMQ rocketmq.ExampleInterface
}

type HelloWorld struct {
	Options
}

// NewHelloWorldService
func NewHelloWorldService(options Options) *HelloWorld {
	return &HelloWorld{
		Options: options,
	}
}

func (s *HelloWorld) SayHello(ctx context.Context, req *helloworldv1.SayHelloRequest) (*helloworldv1.SayHelloResponse, error) {
	xlog.L(ctx).Info("SayHello started", zap.String("name", req.GetName()))

	if req.GetName() == "" {
		return &helloworldv1.SayHelloResponse{
			Error: uint32(helloworldv1.XERROR_ERROR_NAME_EMPTY.GetEcode()),
			Msg:   helloworldv1.XERROR_ERROR_NAME_EMPTY.GetMsg(),
		}, nil
	}

	err := req.Validate()
	if err != nil {
		return &helloworldv1.SayHelloResponse{
			Error: uint32(xerror.InvalidArgument.GetEcode()),
			Msg:   err.Error(),
		}, nil
	}

	err = s.ExampleMysql.Migrate(ctx)
	if err != nil {
		return nil, xerror.Internal
	}

	resp := &helloworldv1.SayHelloResponse_Data{
		Name: "hello " + req.GetName(),
	}

	if req.Name != "done" {
		resp, err := s.ExampleGrpc.SayHello(ctx, &helloworldv1.SayHelloRequest{
			Name: "done",
		})
		if err != nil {
			xlog.L(ctx).Error("ExampleGrpc.SayHello failed", zap.Error(err), zap.Any("res", resp), zap.Any("req", req))
			// return nil, err
		}
		_, err = s.ExampleRedis.Info(ctx)
		if err != nil {
			xlog.L(ctx).Error("ExampleRedis.Info failed", zap.Error(err), zap.Any("res", resp), zap.Any("req", req))
			return nil, xerror.Internal
		}
		_, err = s.ExampleResty.SayHello(ctx)
		if err != nil {
			xlog.L(ctx).Error("ExampleResty.SayHello failed", zap.Error(err), zap.Any("res", resp), zap.Any("req", req))
			// return nil, err
		}
	}

	err = s.ExampleRocketMQ.PushExampleMessage(ctx)
	if err != nil {
		return nil, xerror.Internal
	}

	return &helloworldv1.SayHelloResponse{Data: resp}, nil
}

func (s *HelloWorld) SayHi(ctx context.Context, req *helloworldv1.SayHiRequest) (*helloworldv1.SayHiResponse, error) {
	err := req.Validate()
	if err != nil {
		return &helloworldv1.SayHiResponse{
			Error: uint32(xerror.InvalidArgument.GetEcode()),
			Msg:   err.Error(),
		}, nil
	}

	return &helloworldv1.SayHiResponse{}, nil
}

func (s *HelloWorld) ProcessConsumer(ctx context.Context, msg *primitive.MessageExt) error {
	return nil
}
