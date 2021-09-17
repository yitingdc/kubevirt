/*
 * This file is part of the KubeVirt project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2021 Red Hat, Inc.
 *
 */

package tests_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	k8sres "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"

	v1 "kubevirt.io/client-go/api/v1"
	"kubevirt.io/client-go/kubecli"
	"kubevirt.io/kubevirt/tests"
	cd "kubevirt.io/kubevirt/tests/containerdisk"
	"kubevirt.io/kubevirt/tests/util"
)

var _ = Describe("[sig-compute]Dry-Run requests", func() {
	var err error
	var virtClient kubecli.KubevirtClient
	var restClient *rest.RESTClient

	BeforeEach(func() {
		virtClient, err = kubecli.GetKubevirtClient()
		util.PanicOnError(err)
		restClient = virtClient.RestClient()
		tests.BeforeTestCleanup()
	})

	Context("VirtualMachineInstances", func() {
		var vmi *v1.VirtualMachineInstance
		resource := "virtualmachineinstances"

		BeforeEach(func() {
			vmi = tests.NewRandomVMIWithEphemeralDisk(cd.ContainerDiskFor(cd.ContainerDiskAlpine))
		})

		It("create a VirtualMachineInstance", func() {
			By("Make a Dry-Run request to create a Virtual Machine")
			err = tests.DryRunCreate(restClient, resource, vmi.Namespace, vmi, nil)
			Expect(err).To(BeNil())

			By("Check that no Virtual Machine was actually created")
			_, err = virtClient.VirtualMachineInstance(vmi.Namespace).Get(vmi.Name, &metav1.GetOptions{})
			Expect(errors.IsNotFound(err)).To(BeTrue())
		})

		It("delete a VirtualMachineInstance", func() {
			By("Create a VirtualMachineInstance")
			_, err := virtClient.VirtualMachineInstance(vmi.Namespace).Create(vmi)
			Expect(err).To(BeNil())

			By("Make a Dry-Run request to delete a Virtual Machine")
			deletePolicy := metav1.DeletePropagationForeground
			opts := metav1.DeleteOptions{
				DryRun:            []string{metav1.DryRunAll},
				PropagationPolicy: &deletePolicy,
			}
			err = virtClient.VirtualMachineInstance(vmi.Namespace).Delete(vmi.Name, &opts)
			Expect(err).To(BeNil())

			By("Check that no Virtual Machine was actually deleted")
			_, err = virtClient.VirtualMachineInstance(vmi.Namespace).Get(vmi.Name, &metav1.GetOptions{})
			Expect(err).ToNot(HaveOccurred())
		})

		It("update a VirtualMachineInstance", func() {
			By("Create a VirtualMachineInstance")
			_, err = virtClient.VirtualMachineInstance(vmi.Namespace).Create(vmi)
			Expect(err).To(BeNil())

			By("Make a Dry-Run request to update a Virtual Machine")
			err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
				vmi, err = virtClient.VirtualMachineInstance(vmi.Namespace).Get(vmi.Name, &metav1.GetOptions{})
				if err != nil {
					return err
				}
				vmi.Labels = map[string]string{
					"key": "42",
				}
				return tests.DryRunUpdate(restClient, resource, vmi.Name, vmi.Namespace, vmi, nil)
			})

			By("Check that no update actually took place")
			vmi, err = virtClient.VirtualMachineInstance(vmi.Namespace).Get(vmi.Name, &metav1.GetOptions{})
			Expect(err).ToNot(HaveOccurred())
			Expect(vmi.Labels["key"]).ToNot(Equal("42"))
		})

		It("patch a VirtualMachineInstance", func() {
			By("Create a VirtualMachineInstance")
			vmi, err := virtClient.VirtualMachineInstance(vmi.Namespace).Create(vmi)
			Expect(err).To(BeNil())

			By("Make a Dry-Run request to patch a Virtual Machine")
			patch := []byte(`{"metadata": {"labels": {"key": "42"}}}`)
			err = tests.DryRunPatch(restClient, resource, vmi.Name, vmi.Namespace, types.MergePatchType, patch, nil)
			Expect(err).To(BeNil())

			By("Check that no update actually took place")
			vmi, err = virtClient.VirtualMachineInstance(vmi.Namespace).Get(vmi.Name, &metav1.GetOptions{})
			Expect(err).ToNot(HaveOccurred())
			Expect(vmi.Labels["key"]).ToNot(Equal("42"))
		})
	})

	Context("VirtualMachines", func() {
		var vm *v1.VirtualMachine
		resource := "virtualmachines"

		newVM := func() *v1.VirtualMachine {
			vmiImage := cd.ContainerDiskFor(cd.ContainerDiskCirros)
			vmi := tests.NewRandomVMIWithEphemeralDiskAndUserdata(vmiImage, "echo Hi\n")
			vm := tests.NewRandomVirtualMachine(vmi, false)
			return vm
		}

		BeforeEach(func() {
			vm = newVM()

		})

		It("create a VirtualMachine", func() {
			By("Make a Dry-Run request to create a Virtual Machine")
			err = tests.DryRunCreate(restClient, resource, vm.Namespace, vm, nil)
			Expect(err).To(BeNil())

			By("Check that no Virtual Machine was actually created")
			_, err = virtClient.VirtualMachine(vm.Namespace).Get(vm.Name, &metav1.GetOptions{})
			Expect(errors.IsNotFound(err)).To(BeTrue())
		})

		It("delete a VirtualMachine", func() {
			By("Create a VirtualMachine")
			_, err = virtClient.VirtualMachine(vm.Namespace).Create(vm)
			Expect(err).To(BeNil())

			By("Make a Dry-Run request to delete a Virtual Machine")
			deletePolicy := metav1.DeletePropagationForeground
			opts := metav1.DeleteOptions{
				DryRun:            []string{metav1.DryRunAll},
				PropagationPolicy: &deletePolicy,
			}
			err = virtClient.VirtualMachine(vm.Namespace).Delete(vm.Name, &opts)
			Expect(err).To(BeNil())

			By("Check that no Virtual Machine was actually deleted")
			_, err = virtClient.VirtualMachine(vm.Namespace).Get(vm.Name, &metav1.GetOptions{})
			Expect(err).ToNot(HaveOccurred())
		})

		It("update a VirtualMachine", func() {
			By("Create a VirtualMachine")
			_, err = virtClient.VirtualMachine(vm.Namespace).Create(vm)
			Expect(err).To(BeNil())

			By("Make a Dry-Run request to update a Virtual Machine")
			err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
				vm, err = virtClient.VirtualMachine(vm.Namespace).Get(vm.Name, &metav1.GetOptions{})
				if err != nil {
					return err
				}
				vm.Labels = map[string]string{
					"key": "42",
				}
				return tests.DryRunUpdate(restClient, resource, vm.Name, vm.Namespace, vm, nil)
			})
			Expect(err).To(BeNil())

			By("Check that no update actually took place")
			vm, err = virtClient.VirtualMachine(vm.Namespace).Get(vm.Name, &metav1.GetOptions{})
			Expect(err).ToNot(HaveOccurred())
			Expect(vm.Labels["key"]).ToNot(Equal("42"))
		})

		It("patch a VirtualMachine", func() {
			By("Create a VirtualMachine")
			vm, err = virtClient.VirtualMachine(vm.Namespace).Create(vm)
			Expect(err).To(BeNil())

			By("Make a Dry-Run request to patch a Virtual Machine")
			patch := []byte(`{"metadata": {"labels": {"key": "42"}}}`)
			err = tests.DryRunPatch(restClient, resource, vm.Name, vm.Namespace, types.MergePatchType, patch, nil)
			Expect(err).To(BeNil())

			By("Check that no update actually took place")
			vm, err = virtClient.VirtualMachine(vm.Namespace).Get(vm.Name, &metav1.GetOptions{})
			Expect(err).ToNot(HaveOccurred())
			Expect(vm.Labels["key"]).ToNot(Equal("42"))
		})
	})

	Context("Migrations", func() {
		var vmim *v1.VirtualMachineInstanceMigration
		resource := "virtualmachineinstancemigrations"

		BeforeEach(func() {
			vmi := tests.NewRandomVMIWithEphemeralDisk(cd.ContainerDiskFor(cd.ContainerDiskAlpine))
			vmi, err = virtClient.VirtualMachineInstance(vmi.Namespace).Create(vmi)
			Expect(err).ToNot(HaveOccurred())
			vmim = tests.NewRandomMigration(vmi.Name, vmi.Namespace)
		})

		It("create a migration", func() {
			By("Make a Dry-Run request to create a Migration")
			err = tests.DryRunCreate(restClient, resource, vmim.Namespace, vmim, vmim)
			Expect(err).To(BeNil())

			By("Check that no migration was actually created")
			_, err = virtClient.VirtualMachineInstanceMigration(vmim.Namespace).Get(vmim.Name, &metav1.GetOptions{})
			Expect(errors.IsNotFound(err)).To(BeTrue())
		})

		It("delete a migration", func() {
			By("Create a migration")
			vmim, err = virtClient.VirtualMachineInstanceMigration(vmim.Namespace).Create(vmim)
			Expect(err).To(BeNil())

			By("Make a Dry-Run request to delete a Migration")
			deletePolicy := metav1.DeletePropagationForeground
			opts := metav1.DeleteOptions{
				DryRun:            []string{metav1.DryRunAll},
				PropagationPolicy: &deletePolicy,
			}
			err = virtClient.VirtualMachineInstanceMigration(vmim.Namespace).Delete(vmim.Name, &opts)
			Expect(err).To(BeNil())

			By("Check that no migration was actually deleted")
			_, err = virtClient.VirtualMachineInstanceMigration(vmim.Namespace).Get(vmim.Name, &metav1.GetOptions{})
			Expect(err).ToNot(HaveOccurred())
		})

		It("update a migration", func() {
			By("Create a migration")
			vmim, err := virtClient.VirtualMachineInstanceMigration(vmim.Namespace).Create(vmim)
			Expect(err).To(BeNil())

			By("Make a Dry-Run request to update the migration")
			err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
				vmim, err = virtClient.VirtualMachineInstanceMigration(vmim.Namespace).Get(vmim.Name, &metav1.GetOptions{})
				if err != nil {
					return err
				}
				vmim.Annotations = map[string]string{
					"key": "42",
				}
				return tests.DryRunUpdate(restClient, resource, vmim.Name, vmim.Namespace, vmim, nil)

			})

			Expect(err).ToNot(HaveOccurred())

			By("Check that no update actually took place")
			vmim, err = virtClient.VirtualMachineInstanceMigration(vmim.Namespace).Get(vmim.Name, &metav1.GetOptions{})
			Expect(err).ToNot(HaveOccurred())
			Expect(vmim.Annotations["key"]).ToNot(Equal("42"))
		})

		It("patch a migration", func() {
			By("Create a migration")
			vmim, err = virtClient.VirtualMachineInstanceMigration(vmim.Namespace).Create(vmim)
			Expect(err).To(BeNil())

			By("Make a Dry-Run request to patch the migration")
			patch := []byte(`{"metadata": {"labels": {"key": "42"}}}`)
			err = tests.DryRunPatch(restClient, resource, vmim.Name, vmim.Namespace, types.MergePatchType, patch, nil)
			Expect(err).To(BeNil())

			By("Check that no update actually took place")
			vmim, err = virtClient.VirtualMachineInstanceMigration(vmim.Namespace).Get(vmim.Name, &metav1.GetOptions{})
			Expect(err).ToNot(HaveOccurred())
			Expect(vmim.Labels["key"]).ToNot(Equal("42"))
		})
	})

	Context("VMI Presets", func() {
		var preset *v1.VirtualMachineInstancePreset
		resource := "virtualmachineinstancepresets"
		presetLabelKey := "kubevirt.io/vmi-preset-test"
		presetLabelVal := "test"

		BeforeEach(func() {
			preset = newVMIPreset("test-vmi-preset", presetLabelKey, presetLabelVal)
		})

		It("create a VMI preset", func() {
			By("Make a Dry-Run request to create a VMI preset")
			err = tests.DryRunCreate(restClient, resource, preset.Namespace, preset, nil)
			Expect(err).To(BeNil())

			By("Check that no VMI preset was actually created")
			_, err = virtClient.VirtualMachineInstancePreset(preset.Namespace).Get(preset.Name, metav1.GetOptions{})
			Expect(errors.IsNotFound(err)).To(BeTrue())
		})

		It("delete a VMI preset", func() {
			By("Create a VMI preset")
			_, err := virtClient.VirtualMachineInstancePreset(preset.Namespace).Create(preset)
			Expect(err).To(BeNil())

			By("Make a Dry-Run request to delete a VMI preset")
			deletePolicy := metav1.DeletePropagationForeground
			opts := metav1.DeleteOptions{
				DryRun:            []string{metav1.DryRunAll},
				PropagationPolicy: &deletePolicy,
			}
			err = virtClient.VirtualMachineInstancePreset(preset.Namespace).Delete(preset.Name, &opts)
			Expect(err).To(BeNil())

			By("Check that no VMI preset was actually deleted")
			_, err = virtClient.VirtualMachineInstancePreset(preset.Namespace).Get(preset.Name, metav1.GetOptions{})
			Expect(err).ToNot(HaveOccurred())
		})

		It("update a VMI preset", func() {
			By("Create a VMI preset")
			_, err = virtClient.VirtualMachineInstancePreset(preset.Namespace).Create(preset)
			Expect(err).To(BeNil())

			By("Make a Dry-Run request to update a VMI preset")
			err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
				preset, err = virtClient.VirtualMachineInstancePreset(preset.Namespace).Get(preset.Name, metav1.GetOptions{})
				if err != nil {
					return err
				}

				preset.Labels = map[string]string{
					"key": "42",
				}
				return tests.DryRunUpdate(restClient, resource, preset.Name, preset.Namespace, preset, nil)
			})
			Expect(err).To(BeNil())

			By("Check that no update actually took place")
			preset, err = virtClient.VirtualMachineInstancePreset(preset.Namespace).Get(preset.Name, metav1.GetOptions{})
			Expect(err).ToNot(HaveOccurred())
			Expect(preset.Labels["key"]).ToNot(Equal("42"))
		})

		It("patch a VMI preset", func() {
			By("Create a VMI preset")
			preset, err = virtClient.VirtualMachineInstancePreset(preset.Namespace).Create(preset)
			Expect(err).To(BeNil())

			By("Make a Dry-Run request to patch a VMI preset")
			patch := []byte(`{"metadata": {"labels": {"key": "42"}}}`)
			err = tests.DryRunPatch(restClient, resource, preset.Name, preset.Namespace, types.MergePatchType, patch, nil)
			Expect(err).To(BeNil())

			By("Check that no update actually took place")
			preset, err = virtClient.VirtualMachineInstancePreset(preset.Namespace).Get(preset.Name, metav1.GetOptions{})
			Expect(err).ToNot(HaveOccurred())
			Expect(preset.Labels["key"]).ToNot(Equal("42"))
		})
	})

	Context("VMI ReplicaSets", func() {
		var vmirs *v1.VirtualMachineInstanceReplicaSet
		resource := "virtualmachineinstancereplicasets"

		BeforeEach(func() {
			vmirs = newVMIReplicaSet("test-vmi-rs")
		})

		It("create a VMI replicaset", func() {
			By("Make a Dry-Run request to create a VMI replicaset")
			err = tests.DryRunCreate(restClient, resource, vmirs.Namespace, vmirs, nil)
			Expect(err).To(BeNil())

			By("Check that no VMI replicaset was actually created")
			_, err = virtClient.ReplicaSet(vmirs.Namespace).Get(vmirs.Name, metav1.GetOptions{})
			Expect(errors.IsNotFound(err)).To(BeTrue())
		})

		It("delete a VMI replicaset", func() {
			By("Create a VMI replicaset")
			_, err := virtClient.ReplicaSet(vmirs.Namespace).Create(vmirs)
			Expect(err).To(BeNil())

			By("Make a Dry-Run request to delete a VMI replicaset")
			deletePolicy := metav1.DeletePropagationForeground
			opts := metav1.DeleteOptions{
				DryRun:            []string{metav1.DryRunAll},
				PropagationPolicy: &deletePolicy,
			}
			err = virtClient.ReplicaSet(vmirs.Namespace).Delete(vmirs.Name, &opts)
			Expect(err).To(BeNil())

			By("Check that no VMI replicaset was actually deleted")
			_, err = virtClient.ReplicaSet(vmirs.Namespace).Get(vmirs.Name, metav1.GetOptions{})
			Expect(err).ToNot(HaveOccurred())
		})

		It("update a VMI replicaset", func() {
			By("Create a VMI replicaset")
			_, err = virtClient.ReplicaSet(vmirs.Namespace).Create(vmirs)
			Expect(err).To(BeNil())

			By("Make a Dry-Run request to update a VMI replicaset")
			err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
				vmirs, err = virtClient.ReplicaSet(vmirs.Namespace).Get(vmirs.Name, metav1.GetOptions{})
				if err != nil {
					return err
				}

				vmirs.Labels = map[string]string{
					"key": "42",
				}
				return tests.DryRunUpdate(restClient, resource, vmirs.Name, vmirs.Namespace, vmirs, nil)
			})
			Expect(err).To(BeNil())

			By("Check that no update actually took place")
			vmirs, err = virtClient.ReplicaSet(vmirs.Namespace).Get(vmirs.Name, metav1.GetOptions{})
			Expect(err).ToNot(HaveOccurred())
			Expect(vmirs.Labels["key"]).ToNot(Equal("42"))
		})

		It("patch a VMI replicaset", func() {
			By("Create a VMI replicaset")
			vmirs, err = virtClient.ReplicaSet(vmirs.Namespace).Create(vmirs)
			Expect(err).To(BeNil())

			By("Make a Dry-Run request to patch a VMI replicaset")
			patch := []byte(`{"metadata": {"labels": {"key": "42"}}}`)
			err = tests.DryRunPatch(restClient, resource, vmirs.Name, vmirs.Namespace, types.MergePatchType, patch, nil)
			Expect(err).To(BeNil())

			By("Check that no update actually took place")
			vmirs, err = virtClient.ReplicaSet(vmirs.Namespace).Get(vmirs.Name, metav1.GetOptions{})
			Expect(err).ToNot(HaveOccurred())
			Expect(vmirs.Labels["key"]).ToNot(Equal("42"))
		})
	})
})

func newVMIPreset(name, labelKey, labelValue string) *v1.VirtualMachineInstancePreset {
	return &v1.VirtualMachineInstancePreset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: util.NamespaceTestDefault,
		},
		Spec: v1.VirtualMachineInstancePresetSpec{
			Domain: &v1.DomainSpec{
				Resources: v1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceMemory: k8sres.MustParse("512Mi"),
					},
				},
			},
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					labelKey: labelValue,
				},
			},
		},
	}
}

func newVMIReplicaSet(name string) *v1.VirtualMachineInstanceReplicaSet {
	vmi := tests.NewRandomVMIWithEphemeralDisk(cd.ContainerDiskFor(cd.ContainerDiskAlpine))

	return &v1.VirtualMachineInstanceReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: util.NamespaceTestDefault,
		},
		Spec: v1.VirtualMachineInstanceReplicaSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"kubevirt.io/testrs": "testrs",
				},
			},
			Template: &v1.VirtualMachineInstanceTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"kubevirt.io/testrs": "testrs",
					},
				},
				Spec: vmi.Spec,
			},
		},
	}
}
