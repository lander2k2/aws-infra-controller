package aws

type Infra interface {
	Create() error
	Describe() error
	Delete() error
}

var Inventory struct {
	VpcId             string
	RouteTableId      string
	SubnetId          string
	InternetGatewayId string
	SecurityGroupId   string
	InstanceId        string
}

func Provision(i Infra) error {
	err := i.Create()
	if err != nil {
		return err
	}
	return nil
}

func Get(i Infra) error {
	err := i.Describe()
	if err != nil {
		return err
	}
	return nil
}

func Destroy(i Infra) error {
	err := i.Delete()
	if err != nil {
		return err
	}
	return nil
}
