//
// Copyright 2022 IBM Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package controllers

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	olmv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/types"

	operatorv1alpha1 "github.com/IBM/operand-deployment-lifecycle-manager/api/v1alpha1"
	testdata "github.com/IBM/operand-deployment-lifecycle-manager/controllers/testutil"
)

// +kubebuilder:docs-gen:collapse=Imports

var _ = Describe("OperandConfig controller", func() {
	const (
		name              = "common-service"
		namespace         = "ibm-common-services"
		requestName       = "ibm-cloudpak-name"
		requestNamespace  = "ibm-cloudpak"
		operatorNamespace = "ibm-operators"
	)

	var (
		ctx context.Context

		namespaceName         string
		operatorNamespaceName string
		requestNamespaceName  string

		registry      *operatorv1alpha1.OperandRegistry
		config        *operatorv1alpha1.OperandConfig
		request       *operatorv1alpha1.OperandRequest
		catalogSource *olmv1alpha1.CatalogSource
		configKey     types.NamespacedName
	)

	BeforeEach(func() {
		ctx = context.Background()
		namespaceName = createNSName(namespace)
		operatorNamespaceName = createNSName(operatorNamespace)
		requestNamespaceName = createNSName(requestNamespace)
		registry = testdata.OperandRegistryObj(name, namespaceName, operatorNamespaceName)
		config = testdata.OperandConfigObj(name, namespaceName)
		request = testdata.OperandRequestObj(name, namespaceName, requestName, requestNamespaceName)
		catalogSource = testdata.CatalogSource("community-operators", "openshift-marketplace")
		configKey = types.NamespacedName{Name: name, Namespace: namespaceName}

		By("Creating the Namespace")
		Expect(k8sClient.Create(ctx, testdata.NamespaceObj(namespaceName))).Should(Succeed())
		Expect(k8sClient.Create(ctx, testdata.NamespaceObj(operatorNamespaceName))).Should(Succeed())
		Expect(k8sClient.Create(ctx, testdata.NamespaceObj(requestNamespaceName))).Should(Succeed())
		Expect(k8sClient.Create(ctx, testdata.NamespaceObj("openshift-marketplace")))

		By("Creating the CatalogSource")
		Expect(k8sClient.Create(ctx, catalogSource)).Should(Succeed())
		catalogSource.Status = testdata.CatalogSourceStatus()
		Expect(k8sClient.Status().Update(ctx, catalogSource)).Should(Succeed())
		By("Creating the OperandRegistry")
		Expect(k8sClient.Create(ctx, registry)).Should(Succeed())
		By("Creating the OperandConfig")
		Expect(k8sClient.Create(ctx, config)).Should(Succeed())
	})

	AfterEach(func() {
		By("Deleting the CatalogSource")
		Expect(k8sClient.Delete(ctx, catalogSource)).Should(Succeed())
		By("Deleting the OperandRequest")
		Expect(k8sClient.Delete(ctx, request)).Should(Succeed())
		By("Deleting the OperandConfig")
		Expect(k8sClient.Delete(ctx, config)).Should(Succeed())
		By("Deleting the OperandRegistry")
		Expect(k8sClient.Delete(ctx, registry)).Should(Succeed())
	})

	Context("Initializing OperandConfig Status", func() {
		It("Should the status of OperandConfig be Running", func() {

			By("Checking status of the OperandConfig")
			Eventually(func() operatorv1alpha1.ServicePhase {
				configInstance := &operatorv1alpha1.OperandConfig{}
				Expect(k8sClient.Get(ctx, configKey, configInstance)).Should(Succeed())

				return configInstance.Status.Phase
			}, timeout, interval).Should(Equal(operatorv1alpha1.ServiceInit))

			By("Creating the OperandRequest")
			Expect(k8sClient.Create(ctx, request)).Should(Succeed())

			By("Setting status of the Subscriptions")
			etcdSub := testdata.Subscription("etcd", operatorNamespaceName)
			Eventually(func() error {
				k8sClient.Get(ctx, types.NamespacedName{Name: "etcd", Namespace: operatorNamespaceName}, etcdSub)
				etcdSub.Status = testdata.SubscriptionStatus("etcd", operatorNamespaceName, "0.0.1")
				return k8sClient.Status().Update(ctx, etcdSub)
			}, timeout, interval).Should(Succeed())

			jenkinsSub := testdata.Subscription("jenkins", operatorNamespaceName)
			Eventually(func() error {
				k8sClient.Get(ctx, types.NamespacedName{Name: "jenkins", Namespace: operatorNamespaceName}, jenkinsSub)
				jenkinsSub.Status = testdata.SubscriptionStatus("jenkins", operatorNamespaceName, "0.0.1")
				return k8sClient.Status().Update(ctx, jenkinsSub)
			}, timeout, interval).Should(Succeed())

			By("Creating and Setting status of the ClusterServiceVersions")
			etcdCSV := testdata.ClusterServiceVersion("etcd-csv.v0.0.1", operatorNamespaceName, testdata.EtcdExample)
			Expect(k8sClient.Create(ctx, etcdCSV)).Should(Succeed())
			Eventually(func() error {
				k8sClient.Get(ctx, types.NamespacedName{Name: "etcd-csv.v0.0.1", Namespace: operatorNamespaceName}, etcdCSV)
				etcdCSV.Status = testdata.ClusterServiceVersionStatus()
				return k8sClient.Status().Update(ctx, etcdCSV)
			}, timeout, interval).Should(Succeed())

			jenkinsCSV := testdata.ClusterServiceVersion("jenkins-csv.v0.0.1", operatorNamespaceName, testdata.JenkinsExample)
			Expect(k8sClient.Create(ctx, jenkinsCSV)).Should(Succeed())
			Eventually(func() error {
				k8sClient.Get(ctx, types.NamespacedName{Name: "jenkins-csv.v0.0.1", Namespace: operatorNamespaceName}, jenkinsCSV)
				jenkinsCSV.Status = testdata.ClusterServiceVersionStatus()
				return k8sClient.Status().Update(ctx, jenkinsCSV)
			}, timeout, interval).Should(Succeed())

			By("Creating and Setting status of the InstallPlan")
			etcdIP := testdata.InstallPlan("etcd-install-plan", operatorNamespaceName)
			Expect(k8sClient.Create(ctx, etcdIP)).Should(Succeed())
			Eventually(func() error {
				k8sClient.Get(ctx, types.NamespacedName{Name: "etcd-install-plan", Namespace: operatorNamespaceName}, etcdIP)
				etcdIP.Status = testdata.InstallPlanStatus()
				return k8sClient.Status().Update(ctx, etcdIP)
			}, timeout, interval).Should(Succeed())

			jenkinsIP := testdata.InstallPlan("jenkins-install-plan", operatorNamespaceName)
			Expect(k8sClient.Create(ctx, jenkinsIP)).Should(Succeed())
			Eventually(func() error {
				k8sClient.Get(ctx, types.NamespacedName{Name: "jenkins-install-plan", Namespace: operatorNamespaceName}, jenkinsIP)
				jenkinsIP.Status = testdata.InstallPlanStatus()
				return k8sClient.Status().Update(ctx, jenkinsIP)
			}, timeout, interval).Should(Succeed())

			By("Checking status of the OperandConfig")
			Eventually(func() operatorv1alpha1.ServicePhase {
				configInstance := &operatorv1alpha1.OperandConfig{}
				Expect(k8sClient.Get(ctx, configKey, configInstance)).Should(Succeed())
				return configInstance.Status.Phase
			}, timeout, interval).Should(Equal(operatorv1alpha1.ServiceRunning))

			By("Cleaning up olm resources")
			Expect(k8sClient.Delete(ctx, etcdSub)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, jenkinsSub)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, etcdCSV)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, jenkinsCSV)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, etcdIP)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, jenkinsIP)).Should(Succeed())
		})
	})
})