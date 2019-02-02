package aws

type Infra interface {
	Create() error
	Describe() error
	Delete() error
}

var Config struct {
	ClusterName  string `yaml:"name"`
	Region       string `yaml:"region"`
	MachineImage string `yaml:"ami"`
	KeyName      string `yaml:"keyName"`
}

var Inventory struct {
	Region            string
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
