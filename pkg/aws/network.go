package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type Vpc struct {
	Id     string
	Region string
	Cidr   string
}

type RouteTable struct {
	VpcId  string
	Id     string
	Region string
}

type Subnet struct {
	VpcId  string
	Id     string
	Region string
	Cidr   string
}

type InternetGateway struct {
	VpcId        string
	RouteTableId string
	Id           string
	Region       string
}

type SecurityGroup struct {
	VpcId       string
	Id          string
	Region      string
	GroupName   string
	Description string
}

func (vpc *Vpc) Create() error {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(vpc.Region)}))

	reply, err := svc.CreateVpc(&ec2.CreateVpcInput{
		CidrBlock: aws.String(vpc.Cidr),
	})
	if err != nil {
		return err
	}

	vpc.Id = *reply.Vpc.VpcId

	return nil
}

func (vpc *Vpc) Describe() error {
	return nil
}

func (vpc *Vpc) List() error {
	return nil
}

func (vpc *Vpc) Delete() error {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(vpc.Region)}))

	if _, err := svc.DeleteVpc(&ec2.DeleteVpcInput{
		VpcId: aws.String(vpc.Id),
	}); err != nil {
		return err
	}

	return nil
}

func (rt *RouteTable) Create() error {
	return nil
}

func (rt *RouteTable) Describe() error {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(rt.Region)}))

	reply, err := svc.DescribeRouteTables(&ec2.DescribeRouteTablesInput{
		Filters: []*ec2.Filter{&ec2.Filter{
			Name: aws.String("vpc-id"),
			Values: []*string{
				aws.String(rt.VpcId),
			},
		}},
	})
	if err != nil {
		return err
	}

	rt.Id = *reply.RouteTables[0].RouteTableId

	return nil
}

func (rt *RouteTable) List() error {
	return nil
}

func (rt *RouteTable) Delete() error {
	return nil
}

func (subnet *Subnet) Create() error {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(subnet.Region)}))

	reply, err := svc.CreateSubnet(&ec2.CreateSubnetInput{
		VpcId:     aws.String(subnet.VpcId),
		CidrBlock: aws.String(subnet.Cidr),
	})
	if err != nil {
		return err
	}

	subnet.Id = *reply.Subnet.SubnetId

	return nil
}

func (subnet *Subnet) Describe() error {
	return nil
}

func (subnet *Subnet) List() error {
	return nil
}

func (subnet *Subnet) Delete() error {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(subnet.Region)}))

	if _, err := svc.DeleteSubnet(&ec2.DeleteSubnetInput{
		SubnetId: aws.String(subnet.Id),
	}); err != nil {
		return err
	}

	return nil
}

func (igw *InternetGateway) Create() error {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(igw.Region)}))

	reply, err := svc.CreateInternetGateway(&ec2.CreateInternetGatewayInput{})
	if err != nil {
		return err
	}

	igw.Id = *reply.InternetGateway.InternetGatewayId

	if _, err := svc.AttachInternetGateway(&ec2.AttachInternetGatewayInput{
		InternetGatewayId: reply.InternetGateway.InternetGatewayId,
		VpcId:             aws.String(igw.VpcId),
	}); err != nil {
		return err
	}

	if _, err := svc.CreateRoute(&ec2.CreateRouteInput{
		DestinationCidrBlock: aws.String("0.0.0.0/0"),
		GatewayId:            reply.InternetGateway.InternetGatewayId,
		RouteTableId:         aws.String(igw.RouteTableId),
	}); err != nil {
		return err
	}

	return nil
}

func (igw *InternetGateway) Describe() error {
	return nil
}

func (igw *InternetGateway) List() error {
	return nil
}

func (igw *InternetGateway) Delete() error {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(igw.Region)}))

	if _, err := svc.DetachInternetGateway(&ec2.DetachInternetGatewayInput{
		InternetGatewayId: aws.String(igw.Id),
		VpcId:             aws.String(igw.VpcId),
	}); err != nil {
		return err
	}

	if _, err := svc.DeleteInternetGateway(&ec2.DeleteInternetGatewayInput{
		InternetGatewayId: aws.String(igw.Id),
	}); err != nil {
		return err
	}

	return nil
}

func (sg *SecurityGroup) Create() error {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(sg.Region)}))

	reply, err := svc.CreateSecurityGroup(&ec2.CreateSecurityGroupInput{
		GroupName:   aws.String(sg.GroupName),
		Description: aws.String(sg.Description),
		VpcId:       aws.String(sg.VpcId),
	})
	if err != nil {
		return err
	}

	sg.Id = *reply.GroupId

	if _, err := svc.AuthorizeSecurityGroupIngress(&ec2.AuthorizeSecurityGroupIngressInput{
		GroupId: aws.String(sg.Id),
		IpPermissions: []*ec2.IpPermission{
			(&ec2.IpPermission{}).
				SetIpProtocol("tcp").
				SetFromPort(6443).
				SetToPort(6443).
				SetIpRanges([]*ec2.IpRange{
					{CidrIp: aws.String("0.0.0.0/0")},
				}),
			(&ec2.IpPermission{}).
				SetIpProtocol("tcp").
				SetFromPort(22).
				SetToPort(22).
				SetIpRanges([]*ec2.IpRange{
					{CidrIp: aws.String("0.0.0.0/0")},
				}),
		},
	}); err != nil {
		return err
	}

	return nil
}

func (sg *SecurityGroup) Describe() error {
	return nil
}

func (sg *SecurityGroup) List() error {
	return nil
}

func (sg *SecurityGroup) Delete() error {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(sg.Region)}))

	if _, err := svc.DeleteSecurityGroup(&ec2.DeleteSecurityGroupInput{
		GroupId: aws.String(sg.Id),
	}); err != nil {
		return err
	}

	return nil
}
