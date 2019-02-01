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
	"io/ioutil"
	"log"
	"time"

	"github.com/spf13/cobra"

	"github.com/lander2k2/aws-infra-controller/pkg/aws"
)

// destroyCmd represents the destroy command
var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy an existing cluster",
	Long: `The destroy command will delete all the infrastructure for a specified
cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		inv := aws.Inventory

		log.Print("Collecting inventory...")
		invJson, err := ioutil.ReadFile("/tmp/aws-infra-controller-inventory.json")
		if err != nil {
			log.Print("Failed to read inventory file")
			log.Fatal(err)
		}
		if err := json.Unmarshal(invJson, &inv); err != nil {
			log.Print("Failed to unmarshal inventory json")
			log.Fatal(err)
		}

		log.Print("Deleting EC2 instance...")
		log.Printf("Instance ID: %s", inv.InstanceId)
		instance := aws.Instance{Id: inv.InstanceId}
		if err := aws.Destroy(&instance); err != nil {
			log.Print("Failed to delete instance")
			log.Fatal(err)
		}

		log.Print("Waiting for EC2 instance to terminate...")
		for instance.Status != "terminated" {
			if err := aws.Get(&instance); err != nil {
				log.Print("Failed to get instance")
				log.Fatal(err)
			}
			time.Sleep(time.Second * 2)
			log.Print(".")
		}
		log.Print("EC2 instance terminated")

		log.Print("Deleting security group...")
		log.Printf("Security group ID: %s", inv.SecurityGroupId)
		sg := aws.SecurityGroup{Id: inv.SecurityGroupId}
		if err := aws.Destroy(&sg); err != nil {
			log.Print("Failed to delete security group")
			log.Fatal(err)
		}

		log.Print("Deleting internet gateway...")
		log.Printf("Internet gateway ID: %s", inv.InternetGatewayId)
		igw := aws.InternetGateway{Id: inv.InternetGatewayId, VpcId: inv.VpcId}
		if err := aws.Destroy(&igw); err != nil {
			log.Print("Failed to delete internet gateway")
			log.Fatal(err)
		}

		log.Print("Deleting subnet...")
		log.Printf("Subnet ID: %s", inv.SubnetId)
		subnet := aws.Subnet{Id: inv.SubnetId}
		if err := aws.Destroy(&subnet); err != nil {
			log.Print("Failed to delete subnet")
			log.Fatal(err)
		}

		log.Print("Deleting VPC...")
		log.Printf("VPC ID: %s", inv.VpcId)
		vpc := aws.Vpc{Id: inv.VpcId}
		if err := aws.Destroy(&vpc); err != nil {
			log.Print("Failed to delete VPC")
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// destroyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// destroyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
