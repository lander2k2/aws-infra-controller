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
		log.Print("Creating VPC...")

		vpc := aws.Vpc{Cidr: "10.0.0.0/16"}
		err := aws.Provision(&vpc)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("VPC Id: %s", vpc.Id)

		data := []byte(vpc.Id)
		werr := ioutil.WriteFile("/tmp/aws-infra-controller.txt", data, 0644)
		if err != nil {
			log.Print(werr)
		}

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
