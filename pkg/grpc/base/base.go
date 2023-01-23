package base

type BaseGRPCService struct{}

func (b *BaseGRPCService) Connect() func() {
	// returns connection close function
	return func() {}
}
