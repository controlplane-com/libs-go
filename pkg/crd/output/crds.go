// +build crdgen

package main

import (
	"github.com/controlplane-com/libs-go/pkg/crd"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)
import (
	"fmt"
	"os"

	"github.com/controlplane-com/libs-go/pkg/schema/agent"
	"github.com/controlplane-com/libs-go/pkg/schema/auditctx"
	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/cloudaccount"
	"github.com/controlplane-com/libs-go/pkg/schema/command"
	"github.com/controlplane-com/libs-go/pkg/schema/containerstatus"
	"github.com/controlplane-com/libs-go/pkg/schema/cronjob"
	"github.com/controlplane-com/libs-go/pkg/schema/dbcluster"
	"github.com/controlplane-com/libs-go/pkg/schema/deployment"
	"github.com/controlplane-com/libs-go/pkg/schema/domain"
	"github.com/controlplane-com/libs-go/pkg/schema/group"
	"github.com/controlplane-com/libs-go/pkg/schema/gvc"
	"github.com/controlplane-com/libs-go/pkg/schema/identity"
	"github.com/controlplane-com/libs-go/pkg/schema/image"
	"github.com/controlplane-com/libs-go/pkg/schema/ipSet"
	"github.com/controlplane-com/libs-go/pkg/schema/location"
	"github.com/controlplane-com/libs-go/pkg/schema/memcache"
	"github.com/controlplane-com/libs-go/pkg/schema/mk8s"
	"github.com/controlplane-com/libs-go/pkg/schema/org"
	"github.com/controlplane-com/libs-go/pkg/schema/policy"
	"github.com/controlplane-com/libs-go/pkg/schema/quota"
	"github.com/controlplane-com/libs-go/pkg/schema/secret"
	"github.com/controlplane-com/libs-go/pkg/schema/serviceaccount"
	"github.com/controlplane-com/libs-go/pkg/schema/spicedb"
	"github.com/controlplane-com/libs-go/pkg/schema/tenant"
	"github.com/controlplane-com/libs-go/pkg/schema/user"
	"github.com/controlplane-com/libs-go/pkg/schema/volumeSet"
	"github.com/controlplane-com/libs-go/pkg/schema/workload"
)

func main() {
	var (
		c *apiextensionsv1.CustomResourceDefinition
		yaml string
		err error
	)
	c, err = crd.ConvertStructToCRD(
		&agent.Agent{}, 
		"cpln.io", 
		"v1", 
		"Agent", 
		"agents",
	)
	if err != nil {
		panic(err)
	}
	yaml, err = crd.CRDToYAML(c)
	if err != nil{
		panic(err)
	}
	err = os.WriteFile("./output/yaml/Agent.yaml", []byte(yaml + "\n"), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Wrote CRD YAML for Agent to ./output/yaml/Agent.yaml\n")
	
	c, err = crd.ConvertStructToCRD(
		&auditctx.AuditContext{}, 
		"cpln.io", 
		"v1", 
		"auditctx", 
		"auditcontexts",
	)
	if err != nil {
		panic(err)
	}
	yaml, err = crd.CRDToYAML(c)
	if err != nil{
		panic(err)
	}
	err = os.WriteFile("./output/yaml/AuditContext.yaml", []byte(yaml + "\n"), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Wrote CRD YAML for AuditContext to ./output/yaml/AuditContext.yaml\n")
	
	c, err = crd.ConvertStructToCRD(
		&base.Base{}, 
		"cpln.io", 
		"v1", 
		"Base", 
		"bases",
	)
	if err != nil {
		panic(err)
	}
	yaml, err = crd.CRDToYAML(c)
	if err != nil{
		panic(err)
	}
	err = os.WriteFile("./output/yaml/Base.yaml", []byte(yaml + "\n"), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Wrote CRD YAML for Base to ./output/yaml/Base.yaml\n")
	
	c, err = crd.ConvertStructToCRD(
		&cloudaccount.CloudAccount{}, 
		"cpln.io", 
		"v1", 
		"CloudAccount", 
		"cloudaccounts",
	)
	if err != nil {
		panic(err)
	}
	yaml, err = crd.CRDToYAML(c)
	if err != nil{
		panic(err)
	}
	err = os.WriteFile("./output/yaml/CloudAccount.yaml", []byte(yaml + "\n"), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Wrote CRD YAML for CloudAccount to ./output/yaml/CloudAccount.yaml\n")
	
	c, err = crd.ConvertStructToCRD(
		&command.DeleteVolumeSetSpecVolumeSet{}, 
		"cpln.io", 
		"v1", 
		"DeleteVolumeSetSpecVolumeSet", 
		"deletevolumesetspecvolumesets",
	)
	if err != nil {
		panic(err)
	}
	yaml, err = crd.CRDToYAML(c)
	if err != nil{
		panic(err)
	}
	err = os.WriteFile("./output/yaml/DeleteVolumeSetSpecVolumeSet.yaml", []byte(yaml + "\n"), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Wrote CRD YAML for DeleteVolumeSetSpecVolumeSet to ./output/yaml/DeleteVolumeSetSpecVolumeSet.yaml\n")
	
	c, err = crd.ConvertStructToCRD(
		&containerstatus.ContainerStatus{}, 
		"cpln.io", 
		"v1", 
		"ContainerStatus", 
		"containerstatuses",
	)
	if err != nil {
		panic(err)
	}
	yaml, err = crd.CRDToYAML(c)
	if err != nil{
		panic(err)
	}
	err = os.WriteFile("./output/yaml/ContainerStatus.yaml", []byte(yaml + "\n"), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Wrote CRD YAML for ContainerStatus to ./output/yaml/ContainerStatus.yaml\n")
	
	c, err = crd.ConvertStructToCRD(
		&cronjob.JobExecutionStatus{}, 
		"cpln.io", 
		"v1", 
		"JobExecutionStatus", 
		"jobexecutionstatuses",
	)
	if err != nil {
		panic(err)
	}
	yaml, err = crd.CRDToYAML(c)
	if err != nil{
		panic(err)
	}
	err = os.WriteFile("./output/yaml/JobExecutionStatus.yaml", []byte(yaml + "\n"), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Wrote CRD YAML for JobExecutionStatus to ./output/yaml/JobExecutionStatus.yaml\n")
	
	c, err = crd.ConvertStructToCRD(
		&dbcluster.DbCluster{}, 
		"cpln.io", 
		"v1", 
		"DbCluster", 
		"dbclusters",
	)
	if err != nil {
		panic(err)
	}
	yaml, err = crd.CRDToYAML(c)
	if err != nil{
		panic(err)
	}
	err = os.WriteFile("./output/yaml/DbCluster.yaml", []byte(yaml + "\n"), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Wrote CRD YAML for DbCluster to ./output/yaml/DbCluster.yaml\n")
	
	c, err = crd.ConvertStructToCRD(
		&deployment.Deployment{}, 
		"cpln.io", 
		"v1", 
		"Deployment", 
		"deployments",
	)
	if err != nil {
		panic(err)
	}
	yaml, err = crd.CRDToYAML(c)
	if err != nil{
		panic(err)
	}
	err = os.WriteFile("./output/yaml/Deployment.yaml", []byte(yaml + "\n"), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Wrote CRD YAML for Deployment to ./output/yaml/Deployment.yaml\n")
	
	c, err = crd.ConvertStructToCRD(
		&deployment.DeploymentVersion{}, 
		"cpln.io", 
		"v1", 
		"DeploymentVersion", 
		"deploymentversions",
	)
	if err != nil {
		panic(err)
	}
	yaml, err = crd.CRDToYAML(c)
	if err != nil{
		panic(err)
	}
	err = os.WriteFile("./output/yaml/DeploymentVersion.yaml", []byte(yaml + "\n"), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Wrote CRD YAML for DeploymentVersion to ./output/yaml/DeploymentVersion.yaml\n")
	
	c, err = crd.ConvertStructToCRD(
		&domain.Domain{}, 
		"cpln.io", 
		"v1", 
		"Domain", 
		"domains",
	)
	if err != nil {
		panic(err)
	}
	yaml, err = crd.CRDToYAML(c)
	if err != nil{
		panic(err)
	}
	err = os.WriteFile("./output/yaml/Domain.yaml", []byte(yaml + "\n"), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Wrote CRD YAML for Domain to ./output/yaml/Domain.yaml\n")
	
	c, err = crd.ConvertStructToCRD(
		&group.Group{}, 
		"cpln.io", 
		"v1", 
		"Group", 
		"groups",
	)
	if err != nil {
		panic(err)
	}
	yaml, err = crd.CRDToYAML(c)
	if err != nil{
		panic(err)
	}
	err = os.WriteFile("./output/yaml/Group.yaml", []byte(yaml + "\n"), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Wrote CRD YAML for Group to ./output/yaml/Group.yaml\n")
	
	c, err = crd.ConvertStructToCRD(
		&gvc.Gvc{}, 
		"cpln.io", 
		"v1", 
		"Gvc", 
		"gvcs",
	)
	if err != nil {
		panic(err)
	}
	yaml, err = crd.CRDToYAML(c)
	if err != nil{
		panic(err)
	}
	err = os.WriteFile("./output/yaml/Gvc.yaml", []byte(yaml + "\n"), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Wrote CRD YAML for Gvc to ./output/yaml/Gvc.yaml\n")
	
	c, err = crd.ConvertStructToCRD(
		&identity.Identity{}, 
		"cpln.io", 
		"v1", 
		"Identity", 
		"identities",
	)
	if err != nil {
		panic(err)
	}
	yaml, err = crd.CRDToYAML(c)
	if err != nil{
		panic(err)
	}
	err = os.WriteFile("./output/yaml/Identity.yaml", []byte(yaml + "\n"), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Wrote CRD YAML for Identity to ./output/yaml/Identity.yaml\n")
	
	c, err = crd.ConvertStructToCRD(
		&image.Image{}, 
		"cpln.io", 
		"v1", 
		"Image", 
		"images",
	)
	if err != nil {
		panic(err)
	}
	yaml, err = crd.CRDToYAML(c)
	if err != nil{
		panic(err)
	}
	err = os.WriteFile("./output/yaml/Image.yaml", []byte(yaml + "\n"), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Wrote CRD YAML for Image to ./output/yaml/Image.yaml\n")
	
	c, err = crd.ConvertStructToCRD(
		&ipSet.IpSet{}, 
		"cpln.io", 
		"v1", 
		"IpSet", 
		"ipsets",
	)
	if err != nil {
		panic(err)
	}
	yaml, err = crd.CRDToYAML(c)
	if err != nil{
		panic(err)
	}
	err = os.WriteFile("./output/yaml/IpSet.yaml", []byte(yaml + "\n"), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Wrote CRD YAML for IpSet to ./output/yaml/IpSet.yaml\n")
	
	c, err = crd.ConvertStructToCRD(
		&location.Location{}, 
		"cpln.io", 
		"v1", 
		"Location", 
		"locations",
	)
	if err != nil {
		panic(err)
	}
	yaml, err = crd.CRDToYAML(c)
	if err != nil{
		panic(err)
	}
	err = os.WriteFile("./output/yaml/Location.yaml", []byte(yaml + "\n"), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Wrote CRD YAML for Location to ./output/yaml/Location.yaml\n")
	
	c, err = crd.ConvertStructToCRD(
		&memcache.MemcacheCluster{}, 
		"cpln.io", 
		"v1", 
		"MemcacheCluster", 
		"memcacheclusters",
	)
	if err != nil {
		panic(err)
	}
	yaml, err = crd.CRDToYAML(c)
	if err != nil{
		panic(err)
	}
	err = os.WriteFile("./output/yaml/MemcacheCluster.yaml", []byte(yaml + "\n"), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Wrote CRD YAML for MemcacheCluster to ./output/yaml/MemcacheCluster.yaml\n")
	
	c, err = crd.ConvertStructToCRD(
		&mk8s.Mk8sCluster{}, 
		"cpln.io", 
		"v1", 
		"Mk8sCluster", 
		"mk8sclusters",
	)
	if err != nil {
		panic(err)
	}
	yaml, err = crd.CRDToYAML(c)
	if err != nil{
		panic(err)
	}
	err = os.WriteFile("./output/yaml/Mk8sCluster.yaml", []byte(yaml + "\n"), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Wrote CRD YAML for Mk8sCluster to ./output/yaml/Mk8sCluster.yaml\n")
	
	c, err = crd.ConvertStructToCRD(
		&org.Org{}, 
		"cpln.io", 
		"v1", 
		"Org", 
		"orgs",
	)
	if err != nil {
		panic(err)
	}
	yaml, err = crd.CRDToYAML(c)
	if err != nil{
		panic(err)
	}
	err = os.WriteFile("./output/yaml/Org.yaml", []byte(yaml + "\n"), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Wrote CRD YAML for Org to ./output/yaml/Org.yaml\n")
	
	c, err = crd.ConvertStructToCRD(
		&policy.Policy{}, 
		"cpln.io", 
		"v1", 
		"Policy", 
		"policies",
	)
	if err != nil {
		panic(err)
	}
	yaml, err = crd.CRDToYAML(c)
	if err != nil{
		panic(err)
	}
	err = os.WriteFile("./output/yaml/Policy.yaml", []byte(yaml + "\n"), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Wrote CRD YAML for Policy to ./output/yaml/Policy.yaml\n")
	
	c, err = crd.ConvertStructToCRD(
		&quota.Quota{}, 
		"cpln.io", 
		"v1", 
		"Quota", 
		"quotas",
	)
	if err != nil {
		panic(err)
	}
	yaml, err = crd.CRDToYAML(c)
	if err != nil{
		panic(err)
	}
	err = os.WriteFile("./output/yaml/Quota.yaml", []byte(yaml + "\n"), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Wrote CRD YAML for Quota to ./output/yaml/Quota.yaml\n")
	
	c, err = crd.ConvertStructToCRD(
		&secret.Secret{}, 
		"cpln.io", 
		"v1", 
		"Secret", 
		"secrets",
	)
	if err != nil {
		panic(err)
	}
	yaml, err = crd.CRDToYAML(c)
	if err != nil{
		panic(err)
	}
	err = os.WriteFile("./output/yaml/Secret.yaml", []byte(yaml + "\n"), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Wrote CRD YAML for Secret to ./output/yaml/Secret.yaml\n")
	
	c, err = crd.ConvertStructToCRD(
		&serviceaccount.ServiceAccount{}, 
		"cpln.io", 
		"v1", 
		"ServiceAccount", 
		"serviceaccounts",
	)
	if err != nil {
		panic(err)
	}
	yaml, err = crd.CRDToYAML(c)
	if err != nil{
		panic(err)
	}
	err = os.WriteFile("./output/yaml/ServiceAccount.yaml", []byte(yaml + "\n"), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Wrote CRD YAML for ServiceAccount to ./output/yaml/ServiceAccount.yaml\n")
	
	c, err = crd.ConvertStructToCRD(
		&spicedb.SpicedbCluster{}, 
		"cpln.io", 
		"v1", 
		"SpicedbCluster", 
		"spicedbclusters",
	)
	if err != nil {
		panic(err)
	}
	yaml, err = crd.CRDToYAML(c)
	if err != nil{
		panic(err)
	}
	err = os.WriteFile("./output/yaml/SpicedbCluster.yaml", []byte(yaml + "\n"), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Wrote CRD YAML for SpicedbCluster to ./output/yaml/SpicedbCluster.yaml\n")
	
	c, err = crd.ConvertStructToCRD(
		&tenant.Tenant{}, 
		"cpln.io", 
		"v1", 
		"Tenant", 
		"tenants",
	)
	if err != nil {
		panic(err)
	}
	yaml, err = crd.CRDToYAML(c)
	if err != nil{
		panic(err)
	}
	err = os.WriteFile("./output/yaml/Tenant.yaml", []byte(yaml + "\n"), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Wrote CRD YAML for Tenant to ./output/yaml/Tenant.yaml\n")
	
	c, err = crd.ConvertStructToCRD(
		&user.User{}, 
		"cpln.io", 
		"v1", 
		"User", 
		"users",
	)
	if err != nil {
		panic(err)
	}
	yaml, err = crd.CRDToYAML(c)
	if err != nil{
		panic(err)
	}
	err = os.WriteFile("./output/yaml/User.yaml", []byte(yaml + "\n"), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Wrote CRD YAML for User to ./output/yaml/User.yaml\n")
	
	c, err = crd.ConvertStructToCRD(
		&volumeSet.PersistentVolumeStatus{}, 
		"cpln.io", 
		"v1", 
		"PersistentVolumeStatus", 
		"persistentvolumestatuses",
	)
	if err != nil {
		panic(err)
	}
	yaml, err = crd.CRDToYAML(c)
	if err != nil{
		panic(err)
	}
	err = os.WriteFile("./output/yaml/PersistentVolumeStatus.yaml", []byte(yaml + "\n"), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Wrote CRD YAML for PersistentVolumeStatus to ./output/yaml/PersistentVolumeStatus.yaml\n")
	
	c, err = crd.ConvertStructToCRD(
		&volumeSet.VolumeSet{}, 
		"cpln.io", 
		"v1", 
		"VolumeSet", 
		"volumesets",
	)
	if err != nil {
		panic(err)
	}
	yaml, err = crd.CRDToYAML(c)
	if err != nil{
		panic(err)
	}
	err = os.WriteFile("./output/yaml/VolumeSet.yaml", []byte(yaml + "\n"), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Wrote CRD YAML for VolumeSet to ./output/yaml/VolumeSet.yaml\n")
	
	c, err = crd.ConvertStructToCRD(
		&volumeSet.VolumeSetStatusLocation{}, 
		"cpln.io", 
		"v1", 
		"VolumeSetStatusLocation", 
		"volumesetstatuslocations",
	)
	if err != nil {
		panic(err)
	}
	yaml, err = crd.CRDToYAML(c)
	if err != nil{
		panic(err)
	}
	err = os.WriteFile("./output/yaml/VolumeSetStatusLocation.yaml", []byte(yaml + "\n"), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Wrote CRD YAML for VolumeSetStatusLocation to ./output/yaml/VolumeSetStatusLocation.yaml\n")
	
	c, err = crd.ConvertStructToCRD(
		&workload.Workload{}, 
		"cpln.io", 
		"v1", 
		"Workload", 
		"workloads",
	)
	if err != nil {
		panic(err)
	}
	yaml, err = crd.CRDToYAML(c)
	if err != nil{
		panic(err)
	}
	err = os.WriteFile("./output/yaml/Workload.yaml", []byte(yaml + "\n"), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Wrote CRD YAML for Workload to ./output/yaml/Workload.yaml\n")
	
	fmt.Println("All CRD YAML generated.")
}
