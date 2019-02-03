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
	"log"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	//"syscall"

	"github.com/spf13/cobra"
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

		log.Print("Making directory for kubeconfig file...")
		if err := os.Mkdir("/home/ubuntu/.kube", os.FileMode(0777)); err != nil {
			log.Print("Faild to create kubeconfig directory")
			log.Fatal(err)
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
			"apply", "-f", "/etc/kubernetes/manifests/network.yaml").CombinedOutput()
		if err != nil {
			log.Print("Failed to deploy pod network provider")
			log.Fatal(err)
		}
		log.Print(string(netOut))

		log.Print("Kubernetes cluster booted")
	},
}

func init() {
	rootCmd.AddCommand(bootCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// bootCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// bootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
