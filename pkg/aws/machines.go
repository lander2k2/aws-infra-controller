package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type Instance struct {
	SubnetId        string
	SecurityGroupId string
	Id              string
	ImageId         string
	KeyName         string
	Status          string
}

func (instance *Instance) Create() error {
	// hardcoded var: region ///////////////////////////////////////////////////
	svc := ec2.New(session.New(&aws.Config{Region: aws.String("us-east-2")}))

	reply, err := svc.RunInstances(&ec2.RunInstancesInput{
		ImageId:      aws.String(instance.ImageId),
		InstanceType: aws.String("t2.medium"),
		MinCount:     aws.Int64(1),
		MaxCount:     aws.Int64(1),
		KeyName:      aws.String(instance.KeyName),
		NetworkInterfaces: []*ec2.InstanceNetworkInterfaceSpecification{
			&ec2.InstanceNetworkInterfaceSpecification{
				DeviceIndex:              aws.Int64(0),
				AssociatePublicIpAddress: aws.Bool(true),
				SubnetId:                 aws.String(instance.SubnetId),
				Groups:                   []*string{aws.String(instance.SecurityGroupId)},
				DeleteOnTermination:      aws.Bool(true),
			},
		},
	})
	if err != nil {
		return err
	}

	log.Printf("Reply: %s", reply)
	instance.Id = *reply.Instances[0].InstanceId

	return nil
}

func (instance *Instance) Describe() error {
	// hardcoded var: region ///////////////////////////////////////////////////
	svc := ec2.New(session.New(&aws.Config{Region: aws.String("us-east-2")}))

	reply, err := svc.DescribeInstanceStatus(&ec2.DescribeInstanceStatusInput{
		IncludeAllInstances: aws.Bool(true),
		InstanceIds:         []*string{aws.String(instance.Id)},
	})
	if err != nil {
		return err
	}

	log.Print("Reply: %s", reply)
	instance.Status = *reply.InstanceStatuses[0].InstanceState.Name

	return nil
}

func (instance *Instance) Delete() error {
	// hardcoded var: region ///////////////////////////////////////////////////
	svc := ec2.New(session.New(&aws.Config{Region: aws.String("us-east-2")}))

	reply, err := svc.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: []*string{
			aws.String(instance.Id),
		},
	})
	if err != nil {
		return err
	}

	log.Printf("Reply: %s", reply)

	return nil
}
