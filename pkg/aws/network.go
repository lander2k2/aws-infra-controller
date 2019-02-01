package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type Vpc struct {
	Id   string
	Cidr string
}

func (vpc *Vpc) Create() error {
	// hardcoded var: region ///////////////////////////////////////////////////
	svc := ec2.New(session.New(&aws.Config{Region: aws.String("us-east-2")}))

	reply, err := svc.CreateVpc(&ec2.CreateVpcInput{
		CidrBlock: aws.String(vpc.Cidr),
	})
	if err != nil {
		return err
	}

	log.Printf("Reply: %s", reply)
	vpc.Id = *reply.Vpc.VpcId
	return nil
}

func (vpc *Vpc) Delete() error {
	// hardcoded var: region ///////////////////////////////////////////////////
	svc := ec2.New(session.New(&aws.Config{Region: aws.String("us-east-2")}))

	reply, err := svc.DeleteVpc(&ec2.DeleteVpcInput{
		VpcId: aws.String(vpc.Id),
	})
	if err != nil {
		return err
	}

	log.Printf("Reply: %s", reply)

	return nil
}
