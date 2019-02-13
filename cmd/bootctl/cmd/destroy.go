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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"

	infra "github.com/lander2k2/aws-infra-controller/pkg/apis/infra/v1alpha1"
	"github.com/lander2k2/aws-infra-controller/pkg/aws"
)

var InventoryConfig string

func validateDestroyFlags() error {

	if InventoryConfig == "" {
		return errors.New("'-inventory-config' is a required flag")
	}

	if _, err := os.Stat(InventoryConfig); os.IsNotExist(err) {
		return fmt.Errorf("Inventory config file not found: %s", InventoryConfig)
	}

	return nil
}

// destroyCmd represents the destroy command
var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy an existing cluster",
	Long: `The destroy command will delete all the infrastructure for a specified
cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := validateDestroyFlags(); err != nil {
			log.Print("Failed to validate flags")
			log.Print("'bootctl destroy -h' for help message")
			log.Fatal(err)
		}

		inv := infra.Inventory{}

		log.Print("Reading inventory...")
		invJson, err := ioutil.ReadFile(InventoryConfig)
		if err != nil {
			log.Print("Failed to read inventory file")
			log.Fatal(err)
		}
		if err := json.Unmarshal(invJson, &inv); err != nil {
			log.Print("Failed to unmarshal inventory json")
			log.Fatal(err)
		}

		log.Print("Deleting EC2 instance...")
		log.Printf("Instance ID: %s", inv.Spec.InstanceId)
		instance := aws.Instance{
			Id:     inv.Spec.InstanceId,
			Region: inv.Spec.Region,
		}
		if err := aws.Destroy(&instance); err != nil {
			log.Print("Failed to delete EC2 instance")
			log.Fatal(err)
		}

		log.Print("Waiting for EC2 instance to terminate...")
		for instance.Status != "terminated" {
			if err := aws.Get(&instance); err != nil {
				log.Print("Failed to get instance")
				log.Fatal(err)
			}
			time.Sleep(time.Second * 5)
			log.Print(".")
		}
		log.Print("EC2 instance terminated")

		log.Print("Deleting instance profile...")
		log.Printf("Instance profile ID: %s", inv.Spec.InstanceProfileId)
		profile := aws.InstanceProfile{
			Name:   inv.Spec.InstanceProfileId,
			Region: inv.Spec.Region,
			Role:   inv.Spec.IamRoleId,
		}
		if err := aws.Destroy(&profile); err != nil {
			log.Print("Failed to delete instance profile")
			log.Fatal(err)
		}
		log.Print("Instance profile deleted")

		log.Print("Deleting IAM infra user...")
		log.Printf("User name: %s", inv.Spec.IamUserId)
		user := aws.IamUser{
			Name:        inv.Spec.IamUserId,
			Region:      inv.Spec.Region,
			AccessKeyId: inv.Spec.AccessKeyId,
			Group:       inv.Spec.IamGroupId,
		}
		if err := aws.Destroy(&user); err != nil {
			log.Print("Failed to delete IAM infra user")
			log.Fatal(err)
		}
		log.Print("IAM infra user deleted")

		log.Print("Deleting IAM infra group...")
		log.Printf("Group name: %s", inv.Spec.IamGroupId)
		group := aws.IamGroup{
			Name:   inv.Spec.IamGroupId,
			Region: inv.Spec.Region,
			Policy: inv.Spec.InfraControllerIamPolicyId,
		}
		if err := aws.Destroy(&group); err != nil {
			log.Print("Failed to delete IAM infra group")
			log.Fatal(err)
		}
		log.Print("IAM infra group deleted")

		log.Print("Deleting IAM infra policy...")
		log.Printf("IAM policy ID: %s", inv.Spec.InfraControllerIamPolicyId)
		infraPolicy := aws.IamPolicy{
			Arn:    inv.Spec.InfraControllerIamPolicyId,
			Region: inv.Spec.Region,
		}
		if err := aws.Destroy(&infraPolicy); err != nil {
			log.Print("Failed to delete IAM infra policy")
			log.Fatal(err)
		}
		log.Print("IAM infra policy deleted")

		log.Print("Deleting IAM role...")
		log.Printf("IAM role ID: %s", inv.Spec.IamRoleId)
		role := aws.IamRole{
			Name:   inv.Spec.IamRoleId,
			Policy: inv.Spec.MasterNodeIamPolicyId,
			Region: inv.Spec.Region,
		}
		if err := aws.Destroy(&role); err != nil {
			log.Print("Failed to delete IAM role")
			log.Fatal(err)
		}

		log.Print("Deleting IAM master node policy...")
		log.Printf("IAM policy ID: %s", inv.Spec.MasterNodeIamPolicyId)
		nodePolicy := aws.IamPolicy{
			Arn:    inv.Spec.MasterNodeIamPolicyId,
			Region: inv.Spec.Region,
		}
		if err := aws.Destroy(&nodePolicy); err != nil {
			log.Print("Failed to delete IAM master node policy")
			log.Fatal(err)
		}

		log.Print("Deleting S3 bucket...")
		log.Printf("Bucket name: %s", inv.Spec.BucketId)
		bucket := aws.Bucket{
			Name:   inv.Spec.BucketId,
			Region: inv.Spec.Region,
		}
		if err := aws.Destroy(&bucket); err != nil {
			log.Print("Failed to delete S3 budket")
			log.Fatal(err)
		}

		log.Print("Deleting security group...")
		log.Printf("Security group ID: %s", inv.Spec.SecurityGroupId)
		sg := aws.SecurityGroup{
			Id:     inv.Spec.SecurityGroupId,
			Region: inv.Spec.Region,
		}
		if err := aws.Destroy(&sg); err != nil {
			log.Print("Failed to delete security group")
			log.Fatal(err)
		}

		log.Print("Deleting internet gateway...")
		log.Printf("Internet gateway ID: %s", inv.Spec.InternetGatewayId)
		igw := aws.InternetGateway{
			Id:     inv.Spec.InternetGatewayId,
			VpcId:  inv.Spec.VpcId,
			Region: inv.Spec.Region,
		}
		if err := aws.Destroy(&igw); err != nil {
			log.Print("Failed to delete internet gateway")
			log.Fatal(err)
		}

		log.Print("Deleting subnet...")
		log.Printf("Subnet ID: %s", inv.Spec.SubnetId)
		subnet := aws.Subnet{
			Id:     inv.Spec.SubnetId,
			Region: inv.Spec.Region,
		}
		if err := aws.Destroy(&subnet); err != nil {
			log.Print("Failed to delete subnet")
			log.Fatal(err)
		}

		log.Print("Deleting VPC...")
		log.Printf("VPC ID: %s", inv.Spec.VpcId)
		vpc := aws.Vpc{
			Id:     inv.Spec.VpcId,
			Region: inv.Spec.Region,
		}
		if err := aws.Destroy(&vpc); err != nil {
			log.Print("Failed to delete VPC")
			log.Fatal(err)
		}

		log.Println("Cluster destroyed")
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)

	destroyCmd.Flags().StringVarP(&InventoryConfig, "inventory-config", "i", "", "Inventory configuration file")
}
