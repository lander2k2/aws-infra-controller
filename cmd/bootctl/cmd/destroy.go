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

// destroyCmd represents the destroy command
var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy an existing cluster",
	Long: `The destroy command will delete all the infrastructure for a specified
cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Print("Deleting VPC...")

		data, err := ioutil.ReadFile("/tmp/aws-infra-controller.txt")
		if err != nil {
			log.Fatal(err)
		}

		vpc := aws.Vpc{Id: string(data)}
		derr := aws.Destroy(&vpc)
		if derr != nil {
			log.Fatal(derr)
		}

		log.Printf(string(data))
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
