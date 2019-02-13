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
	"os"
	"os/exec"
	"os/user"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/lander2k2/aws-infra-controller/pkg/aws"
)

var (
	Region    string
	Artifacts string
	Cluster   string
	Machine   string
	Secret    string
)

// bootCmd represents the boot command
var bootCmd = &cobra.Command{
	Use:   "boot",
	Short: "Boot a new Kubernetes master node",
	Long: `The boot command is run on a machine to initialize a new cluster and
start the infrastructure controller.`,
	Run: func(cmd *cobra.Command, args []string) {

		log.Print("Initializing Kubernetes cluster...")
		initOut, err := exec.Command("/usr/bin/kubeadm", "init",
			"--pod-network-cidr=192.168.0.0/16").CombinedOutput()
		if err != nil {
			log.Print("Failed to initialize Kubernetes cluster")
			log.Fatal(err)
		}
		log.Print(string(initOut))

		if _, err := os.Stat("/home/ubuntu/.kube"); os.IsNotExist(err) {
			log.Print("Making directory for kubeconfig file...")
			if err := os.Mkdir("/home/ubuntu/.kube", os.FileMode(0777)); err != nil {
				log.Print("Faild to create kubeconfig directory")
				log.Fatal(err)
			}
		}

		log.Print("Copying kubeconfig file...")
		cpOut, err := exec.Command("/bin/cp", "/etc/kubernetes/admin.conf",
			"/home/ubuntu/.kube/config").CombinedOutput()
		if err != nil {
			log.Print("Failed to copy kubeconfig file")
			log.Fatal(err)
		}
		log.Print(string(cpOut))

		log.Print("Setting ownership for kubeconfig file...")
		user, err := user.Lookup("ubuntu")
		if err != nil {
			log.Print("Failed to lookup ubuntu user")
			log.Fatal(err)
		}
		uid, err := strconv.Atoi(user.Uid)
		if err != nil {
			log.Print("Failed to convert user ID string to int")
			log.Fatal(err)
		}
		gid, err := strconv.Atoi(user.Gid)
		if err != nil {
			log.Print("Failed to convert group ID string to int")
			log.Fatal(err)
		}
		if err := os.Chown("/home/ubuntu/.kube", uid, gid); err != nil {
			log.Print("Failed to set ownership for .kube directory")
			log.Fatal(err)
		}
		if err := os.Chown("/home/ubuntu/.kube/config", uid, gid); err != nil {
			log.Print("Failed to set ownership for kubeconfig file")
			log.Fatal(err)
		}

		log.Print("Deploying pod network provider...")
		netOut, err := exec.Command("/usr/bin/kubectl", "--kubeconfig", "/etc/kubernetes/admin.conf",
			"apply", "-f", "/etc/kubernetes/network/network.yaml").CombinedOutput()
		if err != nil {
			log.Print("Failed to deploy pod network provider")
			log.Fatal(err, string(netOut))
		}
		log.Print(string(netOut))

		log.Print("Create AWS credentials secret...")
		secretCmd := fmt.Sprintf("/bin/echo '%s' | /usr/bin/kubectl --kubeconfig /etc/kubernetes/admin.conf create -f -", Secret)
		secretOut, err := exec.Command("/bin/bash", "-c", secretCmd).CombinedOutput()
		if err != nil {
			log.Print("Failed to create AWS credentials secret")
			log.Fatal(err, string(secretOut))
		}
		log.Print(string(secretOut))

		log.Print("Deploying infra controller...")
		infraOut, err := exec.Command("/usr/bin/kubectl", "--kubeconfig", "/etc/kubernetes/admin.conf",
			"apply", "-f", "/etc/kubernetes/infra/infra.yaml").CombinedOutput()
		if err != nil {
			log.Print("Failed to deploy infra controller")
			log.Fatal(err, string(infraOut))
		}
		log.Print(string(infraOut))

		log.Print("Creating cluster resource...")
		clusterCmd := fmt.Sprintf("/bin/echo '%s' | /usr/bin/kubectl --kubeconfig /etc/kubernetes/admin.conf create -f -", Cluster)
		clusterOut, err := exec.Command("/bin/bash", "-c", clusterCmd).CombinedOutput()
		if err != nil {
			log.Print("Failed to create cluster resource")
			log.Fatal(err, string(clusterOut))
		}
		log.Print(string(clusterOut))

		log.Print("Creating machine resource...")
		machineCmd := fmt.Sprintf("/bin/echo '%s' | /usr/bin/kubectl --kubeconfig /etc/kubernetes/admin.conf create -f -", Machine)
		machineOut, err := exec.Command("/bin/bash", "-c", machineCmd).CombinedOutput()
		if err != nil {
			log.Print("Failed to create machine resource")
			log.Fatal(err, string(machineOut))
		}
		log.Print(string(machineOut))

		log.Print("Creating kubeadm join token...")
		tokenOut, err := exec.Command("/usr/bin/kubeadm", "token", "create", "--print-join-command").CombinedOutput()
		if err != nil {
			log.Print("Failed to create new kubeadm join token")
			log.Fatal(err)
		}
		object := aws.Object{
			Region:   Region,
			Location: Artifacts,
			Body:     string(tokenOut),
		}
		if err := aws.Deposit(&object); err != nil {
			log.Print("Failed to deposit kubeadm join artifact")
			log.Fatal(err)
		}
		log.Print("Join token created and deposited on artifacts store")

		log.Print("Kubernetes cluster booted")
	},
}

func init() {
	rootCmd.AddCommand(bootCmd)

	bootCmd.Flags().StringVarP(&Artifacts, "artifacts", "a", "", "Artifacts store")
	bootCmd.Flags().StringVarP(&Region, "region", "r", "", "AWS region")
	bootCmd.Flags().StringVarP(&Cluster, "cluster", "c", "", "Cluster manifest")
	bootCmd.Flags().StringVarP(&Machine, "machine", "m", "", "Machine manifest")
	bootCmd.Flags().StringVarP(&Secret, "secret", "s", "", "AWS credentials secret")
}
