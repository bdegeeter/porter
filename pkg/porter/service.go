package porter

type ServiceOptions struct {
	Port        int32
	ServiceName string
}

func (o *ServiceOptions) Validate() error {
	return nil
}
