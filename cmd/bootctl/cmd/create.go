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

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

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

		cluster := infra.Cluster{}
		clusterConfig, err := ioutil.ReadFile(ClusterConfig)
		if err != nil {
			log.Print("Failed to read cluster config file")
			log.Fatal(err)
		}
		if err := yaml.Unmarshal(clusterConfig, &cluster); err != nil {
			log.Print("Failed to unmarshal cluster config")
			log.Fatal(err)
		}

		machine := infra.Machine{}
		machineConfig, err := ioutil.ReadFile(MachineConfig)
		if err != nil {
			log.Print("Failed to read machine config file")
			log.Fatal(err)
		}
		if err := yaml.Unmarshal(machineConfig, &machine); err != nil {
			log.Print("Failed to unmarshal machine config")
			log.Fatal(err)
		}

		inv := infra.Inventory{}
		inv.Spec.Region = cluster.Spec.Region

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

		log.Print("Getting route table...")
		rt := aws.RouteTable{
			VpcId:  vpc.Id,
			Region: cluster.Spec.Region,
		}
		if err := aws.Get(&rt); err != nil {
			log.Print("Failed to get route table")
			log.Print("Deleting VPC that was created")
			aws.Destroy(&vpc)
			log.Fatal(err)
		}
		inv.Spec.RouteTableId = rt.Id
		log.Printf("Route table ID: %s", rt.Id)

		log.Print("Creating subnet...")
		subnet := aws.Subnet{
			VpcId:  vpc.Id,
			Region: cluster.Spec.Region,
			Cidr:   "10.0.0.0/18",
		}
		if err := aws.Provision(&subnet); err != nil {
			log.Print("Failed to create subnet")
			log.Print("Deleting VPC that was created")
			aws.Destroy(&vpc)
			log.Fatal(err)
		}
		inv.Spec.SubnetId = subnet.Id
		log.Printf("Subnet ID: %s", subnet.Id)

		log.Print("Creating internet gateway...")
		igw := aws.InternetGateway{
			VpcId:        vpc.Id,
			Region:       cluster.Spec.Region,
			RouteTableId: rt.Id,
		}
		if err := aws.Provision(&igw); err != nil {
			log.Print("Failed to create internet gateway")
			log.Print("Deleting infrastructure that was created")
			aws.Destroy(&subnet)
			aws.Destroy(&vpc)
			log.Fatal(err)
		}
		inv.Spec.InternetGatewayId = igw.Id
		log.Printf("Internet gateway ID: %s", igw.Id)

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
			aws.Destroy(&igw)
			aws.Destroy(&subnet)
			aws.Destroy(&vpc)
			log.Fatal(err)
		}
		inv.Spec.SecurityGroupId = sg.Id
		log.Printf("Security group ID: %s", sg.Id)

		log.Print("Creating EC2 instance...")
		instance := aws.Instance{
			SubnetId:        subnet.Id,
			SecurityGroupId: sg.Id,
			Region:          cluster.Spec.Region,
			ImageId:         machine.Spec.Ami,
			KeyName:         machine.Spec.KeyName,
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
		inv.Spec.InstanceId = instance.Id
		log.Printf("Instance ID: %s", instance.Id)

		invJson, err := json.Marshal(inv)
		if err != nil {
			log.Print("Failed to marshal inventory to json")
			log.Print("Deleting infrastructure that was created")
			aws.Destroy(&instance)
			aws.Destroy(&sg)
			aws.Destroy(&igw)
			aws.Destroy(&subnet)
			aws.Destroy(&vpc)
			log.Fatal(err)
		}

		fmt.Print(string(invJson))
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
	createCmd.Flags().StringVarP(&ClusterConfig, "cluster-config", "c", "", "Cluster configuration file")
	createCmd.Flags().StringVarP(&MachineConfig, "machine-config", "m", "", "Machine configuration file")
}
