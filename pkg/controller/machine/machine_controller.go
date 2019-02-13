/*
Copyright 2019 aws-infra-controller maintainers.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package machine

import (
	"context"
	"encoding/base64"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	infrav1alpha1 "github.com/lander2k2/aws-infra-controller/pkg/apis/infra/v1alpha1"
	"github.com/lander2k2/aws-infra-controller/pkg/aws"
)

var log = logf.Log.WithName("controller")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Machine Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileMachine{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("machine-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Machine
	err = c.Watch(&source.Kind{Type: &infrav1alpha1.Machine{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create
	// Uncomment watch a Deployment created by Machine - change this for objects you create
	//err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
	//	IsController: true,
	//	OwnerType:    &infrav1alpha1.Machine{},
	//})
	//if err != nil {
	//	return err
	//}

	return nil
}

var _ reconcile.Reconciler = &ReconcileMachine{}

// ReconcileMachine reconciles a Machine object
type ReconcileMachine struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Machine object and makes changes based on the state read
// and what is in the Machine.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// a Deployment as an example
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infra.lander2k2.com,resources=machines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infra.lander2k2.com,resources=machines/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infra.lander2k2.com,resources=clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infra.lander2k2.com,resources=clusters/status,verbs=get;update;patch
func (r *ReconcileMachine) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the Machine instance
	machineInstance := &infrav1alpha1.Machine{}
	err := r.Get(context.TODO(), request.NamespacedName, machineInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	log.Info(fmt.Sprintf("Reconciling for machine %s, type %s", machineInstance.ObjectMeta.Name, machineInstance.Spec.MachineType))

	listOpts := client.ListOptions{}
	clusters := &infrav1alpha1.ClusterList{}
	if err := r.List(context.Background(), &listOpts, clusters); err != nil {
		log.Info("Failed to list clusters")
		log.Info(err.Error())
		return reconcile.Result{}, err
	}

	inventoryList := &infrav1alpha1.InventoryList{}
	inv := infrav1alpha1.Inventory{}
	if err := r.List(context.Background(), &listOpts, inventoryList); err != nil {
		log.Info("Failed to list inventory")
		log.Info(err.Error())
		return reconcile.Result{}, err
	}
	if len(inventoryList.Items) < 1 {
		log.Info("No inventory registered")
	} else {
		inv = inventoryList.Items[0]
	}

	switch machineInstance.Spec.MachineType {
	case "boot-master":
		return reconcile.Result{}, nil
	case "worker":
		// get existing worker nodes
		machine := aws.Instance{
			Region:      clusters.Items[0].Spec.Region,
			Cluster:     clusters.Items[0].ObjectMeta.Name,
			MachineType: machineInstance.Spec.MachineType,
		}
		if err := aws.GetAll(&machine); err != nil {
			fmt.Println("Failed to get machines")
			fmt.Println(err)
		}

		switch {
		case int64(machineInstance.Spec.Replicas) > machine.Replicas:
			// provision new instance/s
			count := int64(machineInstance.Spec.Replicas) - machine.Replicas
			userdata := base64.StdEncoding.EncodeToString([]byte(
				fmt.Sprintf("#!/bin/bash\r\nbootctl join -a %s -r %s",
					inv.Spec.BucketId, clusters.Items[0].Spec.Region,
				),
			))
			log.Info("More machines requested")
			instance := aws.Instance{
				SubnetId:        inv.Spec.SubnetId,
				SecurityGroupId: inv.Spec.SecurityGroupId,
				Region:          clusters.Items[0].Spec.Region,
				Cluster:         clusters.Items[0].ObjectMeta.Name,
				ImageId:         machineInstance.Spec.Ami,
				KeyName:         machineInstance.Spec.KeyName,
				Name:            fmt.Sprintf("%s-%s", clusters.Items[0].ObjectMeta.Name, machineInstance.ObjectMeta.Name),
				Profile:         inv.Spec.InstanceProfileId,
				Userdata:        userdata,
				MachineType:     machineInstance.Spec.MachineType,
				Replicas:        count,
			}
			if err := aws.Provision(&instance); err != nil {
				log.Info("Failed to provision instance/s")
				log.Info(err.Error())
			}
			// update inventory

		case int64(machineInstance.Spec.Replicas) < machine.Replicas:
			// destroy instance/s
			log.Info("Fewer machines requested")
			log.Info("Not implemented")
		default:
			log.Info("Desired number of machines running")
		}

		return reconcile.Result{}, nil
	default:
		log.Info(fmt.Sprintf("Do not recognize machine type %s", machineInstance.Spec.MachineType))
		return reconcile.Result{}, nil
	}
}
