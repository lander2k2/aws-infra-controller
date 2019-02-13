// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/ghodss/yaml"
	"github.com/nu7hatch/gouuid"
	"github.com/spf13/cobra"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	infra "github.com/lander2k2/aws-infra-controller/pkg/apis/infra/v1alpha1"
	"github.com/lander2k2/aws-infra-controller/pkg/aws"
)

var (
	ClusterConfig string
	MachineConfig string
)

func validateCreateFlags() error {

	if ClusterConfig == "" {
		return errors.New("'-cluster-config' is a required flag")
	}

	if _, err := os.Stat(ClusterConfig); os.IsNotExist(err) {
		return fmt.Errorf("Cluster config file not found: %s", ClusterConfig)
	}

	if MachineConfig == "" {
		return errors.New("'-machine-config` is a required flag")
	}

	if _, err := os.Stat(MachineConfig); os.IsNotExist(err) {
		return fmt.Errorf("Machine config file not found: %s", MachineConfig)
	}

	return nil
}

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new cluster",
	Long: `The create command will call the infra provider's API and provision the
necessary infrastructure for a new cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := validateCreateFlags(); err != nil {
			log.Print("Failed to validate flags")
			log.Print("'bootctl create -h' for help message")
			log.Fatal(err)
		}

		// load cluster config
		cluster := infra.Cluster{}
		clusterConfigYaml, err := ioutil.ReadFile(ClusterConfig)
		if err != nil {
			log.Print("Failed to read cluster config file")
			log.Fatal(err)
		}
		clusterConfigJson, err := yaml.YAMLToJSON(clusterConfigYaml)
		if err != nil {
			log.Print("Failed to convert cluster config yaml to json")
			log.Fatal(err)
		}
		if err := json.Unmarshal(clusterConfigJson, &cluster); err != nil {
			log.Print("Failed to unmarshal cluster config json")
			log.Fatal(err)
		}

		// load machine config
		machine := infra.Machine{}
		machineConfigYaml, err := ioutil.ReadFile(MachineConfig)
		if err != nil {
			log.Print("Failed to read machine config file")
			log.Fatal(err)
		}
		machineConfigJson, err := yaml.YAMLToJSON(machineConfigYaml)
		if err != nil {
			log.Print("Failed to convert machine config yaml to json")
			log.Fatal(err)
		}
		if err := json.Unmarshal(machineConfigJson, &machine); err != nil {
			log.Print("Failed to unmarshal machine config json")
			log.Fatal(err)
		}

		// create inventory
		inv := infra.Inventory{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Inventory",
				APIVersion: "infra.lander2k2.com/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-inventory", cluster.ObjectMeta.Name),
				Namespace: "kube-system",
			},
		}
		inv.Spec.Region = cluster.Spec.Region

		// create VPC
		log.Print("Creating VPC...")
		vpc := aws.Vpc{
			Cidr:   "10.0.0.0/16",
			Region: cluster.Spec.Region,
		}
		if err := aws.Provision(&vpc); err != nil {
			log.Print("Failed to create VPC")
			log.Fatal(err)
		}
		inv.Spec.VpcId = vpc.Id
		log.Printf("VPC ID: %s", vpc.Id)

		// get route table ID
		log.Print("Getting route table...")
		rt := aws.RouteTable{
			VpcId:  vpc.Id,
			Region: cluster.Spec.Region,
		}
		if err := aws.Get(&rt); err != nil {
			log.Print("Failed to get route table")
			log.Print("Deleting VPC that was created")
			if err := aws.Destroy(&vpc); err != nil {
				log.Print("Failed to delete VPC")
				log.Print(err)
			}
			log.Fatal(err)
		}
		inv.Spec.RouteTableId = rt.Id
		log.Printf("Route table ID: %s", rt.Id)

		// create subnet
		log.Print("Creating subnet...")
		subnet := aws.Subnet{
			VpcId:  vpc.Id,
			Region: cluster.Spec.Region,
			Cidr:   "10.0.0.0/18",
		}
		if err := aws.Provision(&subnet); err != nil {
			log.Print("Failed to create subnet")
			log.Print("Deleting VPC that was created")
			if err := aws.Destroy(&vpc); err != nil {
				log.Print("Failed to delete VPC")
				log.Print(err)
			}
			log.Fatal(err)
		}
		inv.Spec.SubnetId = subnet.Id
		log.Printf("Subnet ID: %s", subnet.Id)

		// create internet gateway
		log.Print("Creating internet gateway...")
		igw := aws.InternetGateway{
			VpcId:        vpc.Id,
			Region:       cluster.Spec.Region,
			RouteTableId: rt.Id,
		}
		if err := aws.Provision(&igw); err != nil {
			log.Print("Failed to create internet gateway")
			log.Print("Deleting infrastructure that was created")
			if err := aws.Destroy(&subnet); err != nil {
				log.Print("Failed to delete subnet")
				log.Print(err)
			}
			if err := aws.Destroy(&vpc); err != nil {
				log.Print("Failed to delete VPC")
				log.Print(err)
			}
			log.Fatal(err)
		}
		inv.Spec.InternetGatewayId = igw.Id
		log.Printf("Internet gateway ID: %s", igw.Id)

		// create security group
		log.Print("Creating security group...")
		sg := aws.SecurityGroup{
			VpcId:       vpc.Id,
			Region:      cluster.Spec.Region,
			GroupName:   fmt.Sprintf("%s-security-group", cluster.ObjectMeta.Name),
			Description: "Kubernetes bootstrap master security group",
		}
		if err := aws.Provision(&sg); err != nil {
			log.Print("Failed to create security group")
			log.Print("Deleting infrastructure that was created")
			if err := aws.Destroy(&igw); err != nil {
				log.Print("Failed to delete internet gateway")
				log.Print(err)
			}
			if err := aws.Destroy(&subnet); err != nil {
				log.Print("Failed to delete subnet")
				log.Print(err)
			}
			if err := aws.Destroy(&vpc); err != nil {
				log.Print("Failed to delete VPC")
				log.Print(err)
			}
			log.Fatal(err)
		}
		inv.Spec.SecurityGroupId = sg.Id
		log.Printf("Security group ID: %s", sg.Id)

		// create S3 bucket
		log.Print("Creating S3 bucket...")
		uuid, err := uuid.NewV4()
		if err != nil {
			log.Print("Failed to create a UUID for S3 bucket name")
			log.Fatal(err)
		}
		bucket := aws.Bucket{
			Region: cluster.Spec.Region,
			Name:   fmt.Sprintf("%s-artifacts-%s", cluster.ObjectMeta.Name, uuid),
		}
		if err := aws.Provision(&bucket); err != nil {
			log.Print("Failed to create s3 bucket")
			log.Print("Deleting infrastructure that was created")
			if err := aws.Destroy(&sg); err != nil {
				log.Print("Failed to delete security group")
				log.Print(err)
			}
			if err := aws.Destroy(&igw); err != nil {
				log.Print("Failed to delete internet gateway")
				log.Print(err)
			}
			if err := aws.Destroy(&subnet); err != nil {
				log.Print("Failed to delete subnet")
				log.Print(err)
			}
			if err := aws.Destroy(&vpc); err != nil {
				log.Print("Failed to delete VPC")
				log.Print(err)
			}
			log.Fatal(err)
		}
		inv.Spec.BucketId = bucket.Name
		log.Printf("S3 bucket name: %s", bucket.Name)

		// create IAM policy for master node
		log.Print("Creating master node IAM policy...")
		policy := aws.IamPolicy{
			Region: cluster.Spec.Region,
			Name:   fmt.Sprintf("%s-node-policy", cluster.ObjectMeta.Name),
			Type:   "machine",
		}
		if err := aws.Provision(&policy); err != nil {
			log.Print("Failed to create master node IAM policy")
			log.Print("Deleting infrastructure that was created")
			if err := aws.Destroy(&bucket); err != nil {
				log.Print("Failed to delete S3 bucket")
				log.Print(err)
			}
			if err := aws.Destroy(&sg); err != nil {
				log.Print("Failed to delete security group")
				log.Print(err)
			}
			if err := aws.Destroy(&igw); err != nil {
				log.Print("Failed to delete internet gateway")
				log.Print(err)
			}
			if err := aws.Destroy(&subnet); err != nil {
				log.Print("Failed to delete subnet")
				log.Print(err)
			}
			if err := aws.Destroy(&vpc); err != nil {
				log.Print("Failed to delete VPC")
				log.Print(err)
			}
			log.Fatal(err)
		}
		inv.Spec.MasterNodeIamPolicyId = policy.Arn
		log.Printf("Master node IAM policy ID: %s", policy.Arn)

		// create IAM role for master node
		log.Print("Creating IAM role...")
		role := aws.IamRole{
			Region: cluster.Spec.Region,
			Name:   fmt.Sprintf("%s-node-role", cluster.ObjectMeta.Name),
			Policy: policy.Arn,
		}
		if err := aws.Provision(&role); err != nil {
			log.Print("Failed to create IAM role")
			log.Print("Deleting infrastructure that was created")
			if err := aws.Destroy(&policy); err != nil {
				log.Print("Failed to delete IAM policy")
				log.Print(err)
			}
			if err := aws.Destroy(&bucket); err != nil {
				log.Print("Failed to delete S3 bucket")
				log.Print(err)
			}
			if err := aws.Destroy(&sg); err != nil {
				log.Print("Failed to delete security group")
				log.Print(err)
			}
			if err := aws.Destroy(&igw); err != nil {
				log.Print("Failed to delete internet gateway")
				log.Print(err)
			}
			if err := aws.Destroy(&subnet); err != nil {
				log.Print("Failed to delete subnet")
				log.Print(err)
			}
			if err := aws.Destroy(&vpc); err != nil {
				log.Print("Failed to delete VPC")
				log.Print(err)
			}
			log.Fatal(err)
		}
		inv.Spec.IamRoleId = role.Name
		log.Printf("IAM role ID: %s", role.Name)

		// create IAM policy for infra controller
		log.Print("Creating infra controller IAM policy...")
		infraPolicy := aws.IamPolicy{
			Region: cluster.Spec.Region,
			Name:   fmt.Sprintf("%s-infra-policy", cluster.ObjectMeta.Name),
			Type:   "infraController",
		}
		if err := aws.Provision(&infraPolicy); err != nil {
			log.Print("Failed to create infra controller IAM policy")
			log.Print("Deleting infrastructure that was created")
			if err := aws.Destroy(&role); err != nil {
				log.Print("Failed to delete IAM role")
				log.Print(err)
			}
			if err := aws.Destroy(&policy); err != nil {
				log.Print("Failed to delete IAM policy")
				log.Print(err)
			}
			if err := aws.Destroy(&bucket); err != nil {
				log.Print("Failed to delete S3 bucket")
				log.Print(err)
			}
			if err := aws.Destroy(&sg); err != nil {
				log.Print("Failed to delete security group")
				log.Print(err)
			}
			if err := aws.Destroy(&igw); err != nil {
				log.Print("Failed to delete internet gateway")
				log.Print(err)
			}
			if err := aws.Destroy(&subnet); err != nil {
				log.Print("Failed to delete subnet")
				log.Print(err)
			}
			if err := aws.Destroy(&vpc); err != nil {
				log.Print("Failed to delete VPC")
				log.Print(err)
			}
			log.Fatal(err)
		}
		inv.Spec.InfraControllerIamPolicyId = infraPolicy.Arn
		log.Printf("Infra IAM policy ID: %s", infraPolicy.Arn)

		// create IAM group and attach policy
		group := aws.IamGroup{
			Region: cluster.Spec.Region,
			Name:   fmt.Sprintf("%s-infra-group", cluster.ObjectMeta.Name),
			Policy: infraPolicy.Arn,
		}
		if err := aws.Provision(&group); err != nil {
			log.Print("Failed to create infra controller group")
			log.Print("Deleting infrastructure that was created")
			if err := aws.Destroy(&infraPolicy); err != nil {
				log.Print("Failed to delete infra controller IAM policy")
				log.Print(err)
			}
			if err := aws.Destroy(&role); err != nil {
				log.Print("Failed to delete IAM role")
				log.Print(err)
			}
			if err := aws.Destroy(&policy); err != nil {
				log.Print("Failed to delete IAM policy")
				log.Print(err)
			}
			if err := aws.Destroy(&bucket); err != nil {
				log.Print("Failed to delete S3 bucket")
				log.Print(err)
			}
			if err := aws.Destroy(&sg); err != nil {
				log.Print("Failed to delete security group")
				log.Print(err)
			}
			if err := aws.Destroy(&igw); err != nil {
				log.Print("Failed to delete internet gateway")
				log.Print(err)
			}
			if err := aws.Destroy(&subnet); err != nil {
				log.Print("Failed to delete subnet")
				log.Print(err)
			}
			if err := aws.Destroy(&vpc); err != nil {
				log.Print("Failed to delete VPC")
				log.Print(err)
			}
			log.Fatal(err)
		}
		inv.Spec.IamGroupId = group.Name
		log.Printf("IAM group name: %s", group.Name)

		// create IAM user and create access key
		user := aws.IamUser{
			Region: cluster.Spec.Region,
			Name:   fmt.Sprintf("%s-infra-user", cluster.ObjectMeta.Name),
			Group:  group.Name,
		}
		if err := aws.Provision(&user); err != nil {
			log.Print("Failed to create infra controller user")
			log.Print("Deleting infrastructure that was created")
			if err := aws.Destroy(&group); err != nil {
				log.Print("Failed to delete IAM group")
				log.Print(err)
			}
			if err := aws.Destroy(&infraPolicy); err != nil {
				log.Print("Failed to delete infra controller IAM policy")
				log.Print(err)
			}
			if err := aws.Destroy(&role); err != nil {
				log.Print("Failed to delete IAM role")
				log.Print(err)
			}
			if err := aws.Destroy(&policy); err != nil {
				log.Print("Failed to delete IAM policy")
				log.Print(err)
			}
			if err := aws.Destroy(&bucket); err != nil {
				log.Print("Failed to delete S3 bucket")
				log.Print(err)
			}
			if err := aws.Destroy(&sg); err != nil {
				log.Print("Failed to delete security group")
				log.Print(err)
			}
			if err := aws.Destroy(&igw); err != nil {
				log.Print("Failed to delete internet gateway")
				log.Print(err)
			}
			if err := aws.Destroy(&subnet); err != nil {
				log.Print("Failed to delete subnet")
				log.Print(err)
			}
			if err := aws.Destroy(&vpc); err != nil {
				log.Print("Failed to delete VPC")
				log.Print(err)
			}
			log.Fatal(err)
		}
		inv.Spec.IamUserId = user.Name
		inv.Spec.AccessKeyId = user.AccessKeyId
		log.Printf("IAM user name: %s", user.Name)

		// create EC2 instance profile
		log.Print("Creating instance profile...")
		profile := aws.InstanceProfile{
			Region: cluster.Spec.Region,
			Name:   fmt.Sprintf("%s-profile", cluster.ObjectMeta.Name),
			Role:   role.Name,
		}
		if err := aws.Provision(&profile); err != nil {
			log.Print("Failed to create instance profile")
			log.Print("Deleting infrastructure that was created")
			if err := aws.Destroy(&role); err != nil {
				log.Print("Failed to delete IAM role")
				log.Print(err)
			}
			if err := aws.Destroy(&policy); err != nil {
				log.Print("Failed to delete IAM policy")
				log.Print(err)
			}
			if err := aws.Destroy(&bucket); err != nil {
				log.Print("Failed to delete S3 bucket")
				log.Print(err)
			}
			if err := aws.Destroy(&sg); err != nil {
				log.Print("Failed to delete security group")
				log.Print(err)
			}
			if err := aws.Destroy(&igw); err != nil {
				log.Print("Failed to delete internet gateway")
				log.Print(err)
			}
			if err := aws.Destroy(&subnet); err != nil {
				log.Print("Failed to delete subnet")
				log.Print(err)
			}
			if err := aws.Destroy(&vpc); err != nil {
				log.Print("Failed to delete VPC")
				log.Print(err)
			}
			log.Fatal(err)
		}
		inv.Spec.InstanceProfileId = profile.Name
		log.Printf("Instance profile ID: %s", profile.Name)

		// instance profile seems to take a few seconds to become available
		log.Print("Pausing to let instance profile become ready...")
		time.Sleep(time.Second * 15)

		// create secret with AWS access credentials for infra controller
		secret := &v1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Secret",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "aws-creds",
				Namespace: "kube-system",
			},
			Data: map[string][]byte{
				"aws_access_key_id":     []byte(user.AccessKeyId),
				"aws_secret_access_key": []byte(user.SecretAccessKey),
			},
		}
		secretJson, err := json.Marshal(secret)
		if err != nil {
			log.Println("Failed to marshal secret json")
			log.Fatal(err)
		}

		log.Print("Creating EC2 instance...")
		userdata := base64.StdEncoding.EncodeToString([]byte(
			fmt.Sprintf("#!/bin/bash\r\nbootctl boot -a %s -r %s -c '%s' -m '%s' -s '%s'",
				bucket.Name, cluster.Spec.Region,
				string(clusterConfigJson), string(machineConfigJson), string(secretJson),
			),
		))
		instance := aws.Instance{
			SubnetId:        subnet.Id,
			SecurityGroupId: sg.Id,
			Region:          cluster.Spec.Region,
			Cluster:         cluster.ObjectMeta.Name,
			ImageId:         machine.Spec.Ami,
			KeyName:         machine.Spec.KeyName,
			Name:            fmt.Sprintf("%s-%s", cluster.ObjectMeta.Name, machine.ObjectMeta.Name),
			Profile:         profile.Name,
			Userdata:        userdata,
			MachineType:     machine.Spec.MachineType,
			Replicas:        1,
		}
		if err := aws.Provision(&instance); err != nil {
			log.Print("Failed to create EC2 instance")
			log.Print("Deleting infrastructure that was created")
			if err := aws.Destroy(&profile); err != nil {
				log.Print("Failed to delete instance profile")
				log.Print(err)
			}
			if err := aws.Destroy(&role); err != nil {
				log.Print("Failed to delete IAM role")
				log.Print(err)
			}
			if err := aws.Destroy(&policy); err != nil {
				log.Print("Failed to delete IAM policy")
				log.Print(err)
			}
			if err := aws.Destroy(&bucket); err != nil {
				log.Print("Failed to delete S3 bucket")
				log.Print(err)
			}
			if err := aws.Destroy(&sg); err != nil {
				log.Print("Failed to delete security group")
				log.Print(err)
			}
			if err := aws.Destroy(&igw); err != nil {
				log.Print("Failed to delete internet gateway")
				log.Print(err)
			}
			if err := aws.Destroy(&subnet); err != nil {
				log.Print("Failed to delete subnet")
				log.Print(err)
			}
			if err := aws.Destroy(&vpc); err != nil {
				log.Print("Failed to delete VPC")
				log.Print(err)
			}
			log.Fatal(err)
		}
		inv.Spec.InstanceId = instance.Id
		log.Printf("Instance ID: %s", instance.Id)

		invJson, err := json.Marshal(inv)
		if err != nil {
			log.Print("Failed to marshal inventory to json")
			log.Print("Deleting infrastructure that was created")
			if err := aws.Destroy(&profile); err != nil {
				log.Print("Failed to delete instance profile")
				log.Print(err)
			}
			if err := aws.Destroy(&role); err != nil {
				log.Print("Failed to delete IAM role")
				log.Print(err)
			}
			if err := aws.Destroy(&bucket); err != nil {
				log.Print("Failed to delete S3 bucket")
				log.Print(err)
			}
			if err := aws.Destroy(&sg); err != nil {
				log.Print("Failed to delete security group")
				log.Print(err)
			}
			if err := aws.Destroy(&igw); err != nil {
				log.Print("Failed to delete internet gateway")
				log.Print(err)
			}
			if err := aws.Destroy(&subnet); err != nil {
				log.Print("Failed to delete subnet")
				log.Print(err)
			}
			if err := aws.Destroy(&vpc); err != nil {
				log.Print("Failed to delete VPC")
				log.Print(err)
			}
			log.Fatal(err)
		}

		fmt.Print(string(invJson))
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().StringVarP(&ClusterConfig, "cluster-config", "c", "", "Cluster configuration file")
	createCmd.Flags().StringVarP(&MachineConfig, "machine-config", "m", "", "Machine configuration file")
}
