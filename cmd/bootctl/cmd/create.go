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
	"fmt"
	"io/ioutil"
	"log"

	"github.com/spf13/cobra"

	"github.com/lander2k2/aws-infra-controller/pkg/aws"
)

var Name string

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new cluster",
	Long: `The create command will call the infra provider's API and provision the
necessary infrastructure for a new cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		inv := aws.Inventory

		log.Print("Creating VPC...")
		vpc := aws.Vpc{Cidr: "10.0.0.0/16"}
		if err := aws.Provision(&vpc); err != nil {
			log.Print("Failed to create VPC")
			log.Fatal(err)
		}
		inv.VpcId = vpc.Id
		log.Printf("VPC ID: %s", vpc.Id)

		log.Print("Getting route table...")
		rt := aws.RouteTable{VpcId: vpc.Id}
		if err := aws.Get(&rt); err != nil {
			log.Print("Failed to get route table")
			log.Print("Deleting VPC that was created")
			aws.Destroy(&vpc)
			log.Fatal(err)
		}
		inv.RouteTableId = rt.Id
		log.Printf("Route table ID: %s", rt.Id)

		log.Print("Creating subnet...")
		subnet := aws.Subnet{VpcId: vpc.Id, Cidr: "10.0.0.0/18"}
		if err := aws.Provision(&subnet); err != nil {
			log.Print("Failed to create subnet")
			log.Print("Deleting VPC that was created")
			aws.Destroy(&vpc)
			log.Fatal(err)
		}
		inv.SubnetId = subnet.Id

		log.Print("Creating internet gateway...")
		igw := aws.InternetGateway{VpcId: vpc.Id, RouteTableId: rt.Id}
		if err := aws.Provision(&igw); err != nil {
			log.Print("Failed to create internet gateway")
			log.Print("Deleting infrastructure that was created")
			aws.Destroy(&subnet)
			aws.Destroy(&vpc)
			log.Fatal(err)
		}
		inv.InternetGatewayId = igw.Id

		log.Print("Creating security group...")
		sg := aws.SecurityGroup{
			VpcId:       vpc.Id,
			GroupName:   fmt.Sprintf("%s-security-group", Name),
			Description: "Kubernetes bootstrap master security group",
		}
		if err := aws.Provision(&sg); err != nil {
			log.Print("Failed to create security group")
			log.Print("Deleting infrastructure that was created")
			aws.Destroy(&igw)
			aws.Destroy(&subnet)
			aws.Destroy(&vpc)
			log.Fatal(err)
		}
		inv.SecurityGroupId = sg.Id

		log.Print("Creating EC2 instance...")
		instance := aws.Instance{
			SubnetId:        subnet.Id,
			SecurityGroupId: sg.Id,
			// hard coded var ami, ssh key /////////////////////////////////////
			ImageId: "ami-0e2a10a0edd037f7e",
			KeyName: "dev-richard",
		}
		if err := aws.Provision(&instance); err != nil {
			log.Print("Failed to create EC2 instance")
			log.Print("Deleting infrastructure that was created")
			aws.Destroy(&sg)
			aws.Destroy(&igw)
			aws.Destroy(&subnet)
			aws.Destroy(&vpc)
			log.Fatal(err)
		}
		inv.InstanceId = instance.Id

		///////////////////////////////////////////////////////////////////////
		invJson, err := json.Marshal(inv)
		if err != nil {
			log.Print("Failed to marshal inventory to json")
			log.Print("Deleting VPC")
			aws.Destroy(&vpc)
			log.Fatal(err)
		}

		invContent := []byte(invJson)
		if err := ioutil.WriteFile("/tmp/aws-infra-controller-inventory.json", invContent, 0644); err != nil {
			log.Print("Failed to write inventory file")
			log.Print("Deleting VPC")
			aws.Destroy(&vpc)
			log.Fatal(err)
		}
		log.Print(string(invJson))
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	createCmd.Flags().StringVarP(&Name, "name", "n", "test-01", "Name for new cluster")
}
