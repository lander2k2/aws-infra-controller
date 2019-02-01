package aws

type Infra interface {
	Create() error
	Delete() error
}

func Provision(i Infra) error {
	err := i.Create()
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
