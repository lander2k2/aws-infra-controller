package aws

import (
	"errors"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
)

type Instance struct {
	SubnetId        string
	SecurityGroupId string
	Id              string
	Name            string
	Region          string
	Cluster         string
	ImageId         string
	KeyName         string
	Status          string
	Profile         string
	Userdata        string
	MachineType     string
	Replicas        int64
}

type IamPolicy struct {
	Region string
	Name   string
	Arn    string
	Type   string
}

type IamRole struct {
	Region string
	Name   string
	Policy string
}

type IamGroup struct {
	Region string
	Name   string
	Policy string
}

type IamUser struct {
	Region          string
	Name            string
	Group           string
	AccessKeyId     string
	SecretAccessKey string
}

type InstanceProfile struct {
	Region string
	Name   string
	Role   string
	Arn    string
}

var (
	machinePolicy = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action" : [
        "s3:GetObject",
        "s3:PutObject"
      ],
      "Resource": "arn:aws:s3:::*",
      "Effect": "Allow"
    }
  ]
}`

	infraControllerPolicy = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action" : "*",
      "Resource": "*",
      "Effect": "Allow"
    }
  ]
}`
)

func (instance *Instance) Create() error {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(instance.Region)}))

	reply, err := svc.RunInstances(&ec2.RunInstancesInput{
		ImageId:      aws.String(instance.ImageId),
		InstanceType: aws.String("t2.medium"),
		MinCount:     aws.Int64(instance.Replicas),
		MaxCount:     aws.Int64(instance.Replicas),
		KeyName:      aws.String(instance.KeyName),
		UserData:     aws.String(instance.Userdata),
		NetworkInterfaces: []*ec2.InstanceNetworkInterfaceSpecification{
			&ec2.InstanceNetworkInterfaceSpecification{
				DeviceIndex:              aws.Int64(0),
				AssociatePublicIpAddress: aws.Bool(true),
				SubnetId:                 aws.String(instance.SubnetId),
				Groups:                   []*string{aws.String(instance.SecurityGroupId)},
				DeleteOnTermination:      aws.Bool(true),
			},
		},
		IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
			Name: aws.String(instance.Profile),
		},
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("instance"),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(instance.Name),
					},
					{
						Key:   aws.String("Cluster"),
						Value: aws.String(instance.Cluster),
					},
					{
						Key:   aws.String("MachineType"),
						Value: aws.String(instance.MachineType),
					},
				},
			},
		},
	})
	if err != nil {
		return err
	}

	instance.Id = *reply.Instances[0].InstanceId

	return nil
}

func (instance *Instance) Describe() error {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(instance.Region)}))

	reply, err := svc.DescribeInstanceStatus(&ec2.DescribeInstanceStatusInput{
		IncludeAllInstances: aws.Bool(true),
		InstanceIds:         []*string{aws.String(instance.Id)},
	})
	if err != nil {
		return err
	}

	instance.Status = *reply.InstanceStatuses[0].InstanceState.Name

	return nil
}

func (instance *Instance) List() error {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(instance.Region)}))

	reply, err := svc.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:Cluster"),
				Values: []*string{
					aws.String(instance.Cluster),
				},
			},
			{
				Name: aws.String("tag:MachineType"),
				Values: []*string{
					aws.String(instance.MachineType),
				},
			},
		},
	})
	if err != nil {
		return err
	}

	replicas := int64(0)
	for _, r := range reply.Reservations {
		for _, i := range r.Instances {
			if *i.State.Code < int64(17) {
				log.Print("Found pending/running instance ", i.InstanceId)
				replicas++
			} else {
				log.Print("Found NOT RUNNING instance ", i.InstanceId)
			}
		}
	}
	instance.Replicas = replicas

	return nil
}

func (instance *Instance) Delete() error {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(instance.Region)}))

	if _, err := svc.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: []*string{
			aws.String(instance.Id),
		},
	}); err != nil {
		return err
	}

	return nil
}

func (policy *IamPolicy) Create() error {
	svc := iam.New(session.New(&aws.Config{Region: aws.String(policy.Region)}))

	var policyDoc string
	switch policy.Type {
	case "machine":
		policyDoc = machinePolicy
	case "infraController":
		policyDoc = infraControllerPolicy
	default:
		return errors.New("Unrecognized policy type")
	}

	reply, err := svc.CreatePolicy(&iam.CreatePolicyInput{
		PolicyName:     aws.String(policy.Name),
		PolicyDocument: aws.String(policyDoc),
	})
	if err != nil {
		return err
	}

	policy.Arn = *reply.Policy.Arn

	return nil
}

func (policy *IamPolicy) Describe() error {
	return nil
}

func (policy *IamPolicy) List() error {
	return nil
}

func (policy *IamPolicy) Delete() error {
	svc := iam.New(session.New(&aws.Config{Region: aws.String(policy.Region)}))

	if _, err := svc.DeletePolicy(&iam.DeletePolicyInput{
		PolicyArn: aws.String(policy.Arn),
	}); err != nil {
		return err
	}

	return nil
}

func (role *IamRole) Create() error {
	svc := iam.New(session.New(&aws.Config{Region: aws.String(role.Region)}))

	assume := `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "sts:AssumeRole",
            "Principal": {
                "Service": "ec2.amazonaws.com"
            },
            "Effect": "Allow",
            "Sid": ""
        }
    ]
}`

	if _, err := svc.CreateRole(&iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String(assume),
		RoleName:                 aws.String(role.Name),
	}); err != nil {
		return err
	}

	if _, err := svc.AttachRolePolicy(&iam.AttachRolePolicyInput{
		PolicyArn: aws.String(role.Policy),
		RoleName:  aws.String(role.Name),
	}); err != nil {
		return err
	}

	return nil
}

func (role *IamRole) Describe() error {
	return nil
}

func (role *IamRole) List() error {
	return nil
}

func (role *IamRole) Delete() error {
	svc := iam.New(session.New(&aws.Config{Region: aws.String(role.Region)}))

	if _, err := svc.DetachRolePolicy(&iam.DetachRolePolicyInput{
		PolicyArn: aws.String(role.Policy),
		RoleName:  aws.String(role.Name),
	}); err != nil {
		return err
	}

	if _, err := svc.DeleteRole(&iam.DeleteRoleInput{
		RoleName: aws.String(role.Name),
	}); err != nil {
		return err
	}

	return nil
}

func (group *IamGroup) Create() error {
	svc := iam.New(session.New(&aws.Config{Region: aws.String(group.Region)}))

	if _, err := svc.CreateGroup(&iam.CreateGroupInput{
		GroupName: aws.String(group.Name),
	}); err != nil {
		return err
	}

	if _, err := svc.AttachGroupPolicy(&iam.AttachGroupPolicyInput{
		GroupName: aws.String(group.Name),
		PolicyArn: aws.String(group.Policy),
	}); err != nil {
		return err
	}

	return nil
}

func (group *IamGroup) Describe() error {
	return nil
}

func (group *IamGroup) List() error {
	return nil
}

func (group *IamGroup) Delete() error {
	svc := iam.New(session.New(&aws.Config{Region: aws.String(group.Region)}))

	if _, err := svc.DetachGroupPolicy(&iam.DetachGroupPolicyInput{
		GroupName: aws.String(group.Name),
		PolicyArn: aws.String(group.Policy),
	}); err != nil {
		return err
	}

	if _, err := svc.DeleteGroup(&iam.DeleteGroupInput{
		GroupName: aws.String(group.Name),
	}); err != nil {
		return err
	}

	return nil
}

func (user *IamUser) Create() error {
	svc := iam.New(session.New(&aws.Config{Region: aws.String(user.Region)}))

	if _, err := svc.CreateUser(&iam.CreateUserInput{
		UserName: aws.String(user.Name),
	}); err != nil {
		return err
	}

	if _, err := svc.AddUserToGroup(&iam.AddUserToGroupInput{
		UserName:  aws.String(user.Name),
		GroupName: aws.String(user.Group),
	}); err != nil {
		return err
	}

	reply, err := svc.CreateAccessKey(&iam.CreateAccessKeyInput{
		UserName: aws.String(user.Name),
	})
	if err != nil {
		return err
	}

	user.AccessKeyId = *reply.AccessKey.AccessKeyId
	user.SecretAccessKey = *reply.AccessKey.SecretAccessKey

	return nil
}

func (user *IamUser) Describe() error {
	return nil
}

func (user *IamUser) List() error {
	return nil
}

func (user *IamUser) Delete() error {
	svc := iam.New(session.New(&aws.Config{Region: aws.String(user.Region)}))

	if _, err := svc.DeleteAccessKey(&iam.DeleteAccessKeyInput{
		UserName:    aws.String(user.Name),
		AccessKeyId: aws.String(user.AccessKeyId),
	}); err != nil {
		return err
	}

	if _, err := svc.RemoveUserFromGroup(&iam.RemoveUserFromGroupInput{
		UserName:  aws.String(user.Name),
		GroupName: aws.String(user.Group),
	}); err != nil {
		return err
	}

	if _, err := svc.DeleteUser(&iam.DeleteUserInput{
		UserName: aws.String(user.Name),
	}); err != nil {
		return err
	}

	return nil
}

func (profile *InstanceProfile) Create() error {
	svc := iam.New(session.New(&aws.Config{Region: aws.String(profile.Region)}))

	reply, err := svc.CreateInstanceProfile(&iam.CreateInstanceProfileInput{
		InstanceProfileName: aws.String(profile.Name),
	})
	if err != nil {
		return err
	}

	profile.Arn = *reply.InstanceProfile.Arn

	if _, err := svc.AddRoleToInstanceProfile(&iam.AddRoleToInstanceProfileInput{
		InstanceProfileName: aws.String(profile.Name),
		RoleName:            aws.String(profile.Role),
	}); err != nil {
		return err
	}

	return nil
}

func (profile *InstanceProfile) Describe() error {
	svc := iam.New(session.New(&aws.Config{Region: aws.String(profile.Region)}))

	if _, err := svc.GetInstanceProfile(&iam.GetInstanceProfileInput{
		InstanceProfileName: aws.String(profile.Name),
	}); err != nil {
		return err
	}

	return nil
}

func (profile *InstanceProfile) List() error {
	return nil
}

func (profile *InstanceProfile) Delete() error {
	svc := iam.New(session.New(&aws.Config{Region: aws.String(profile.Region)}))

	if _, err := svc.RemoveRoleFromInstanceProfile(&iam.RemoveRoleFromInstanceProfileInput{
		InstanceProfileName: aws.String(profile.Name),
		RoleName:            aws.String(profile.Role),
	}); err != nil {
		return err
	}

	if _, err := svc.DeleteInstanceProfile(&iam.DeleteInstanceProfileInput{
		InstanceProfileName: aws.String(profile.Name),
	}); err != nil {
		return err
	}

	return nil
}
