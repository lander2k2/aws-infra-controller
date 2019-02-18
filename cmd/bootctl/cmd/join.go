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
	"fmt"
	"log"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/lander2k2/aws-infra-controller/pkg/aws"
)

var (
	JoinRegion    string
	JoinArtifacts string
)

// joinCmd represents the join command
var joinCmd = &cobra.Command{
	Use:   "join",
	Short: "Join a worker node to a Kubernetes cluster",
	Long: `The join command is run on a new worker node to join it to a Kubernetes
cluster using kubeadm and a join command retrieved from an artifacts store.`,
	Run: func(cmd *cobra.Command, args []string) {

		log.Print("Retrieving join command from artifact store...")
		object := aws.Object{
			Region:   JoinRegion,
			Location: JoinArtifacts,
		}
		if err := aws.Retrieve(&object); err != nil {
			log.Print("Failed to retrieve kubeadm join artifact")
			log.Fatal(err)
		}
		log.Print("Join token retrieved")

		log.Print("Joining node to Kubernetes cluster...")
		joinCmd := fmt.Sprintf("/usr/bin/" + object.Body)
		joinOut, err := exec.Command("/bin/bash", "-c", joinCmd).CombinedOutput()
		if err != nil {
			log.Print("Failed to join node to Kubernetes cluster")
			log.Fatal(err)
		}
		log.Print(string(joinOut))
	},
}

func init() {
	rootCmd.AddCommand(joinCmd)

	joinCmd.Flags().StringVarP(&JoinArtifacts, "artifacts", "a", "", "Artifacts store")
	joinCmd.Flags().StringVarP(&JoinRegion, "region", "r", "", "AWS region")
}
