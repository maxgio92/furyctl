// Copyright (c) 2020 SIGHUP s.r.l All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vsphere

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobuffalo/packr/v2"
	getter "github.com/hashicorp/go-getter"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/sighupio/furyagent/pkg/component"
	"github.com/sighupio/furyagent/pkg/storage"
	cfg "github.com/sighupio/furyctl/internal/cluster/configuration"
	"github.com/sighupio/furyctl/internal/configuration"
	log "github.com/sirupsen/logrus"
)

// VSphere represents the VSphere provisioner
type VSphere struct {
	terraform *tfexec.Terraform
	box       *packr.Box
	config    *configuration.Configuration
}

// InitMessage return a custom provisioner message the user will see once the cluster is ready to be updated
func (e *VSphere) InitMessage() string {
	return `[VSphere] Fury

TODO: ...
`
}

// UpdateMessage return a custom provisioner message the user will see once the cluster is updated
func (e *VSphere) UpdateMessage() string {
	// Take the output from Terraform
	// Take the output from Ansible
	// Format everything in a nice way
	return "TODO"
}

// DestroyMessage return a custom provisioner message the user will see once the cluster is destroyed
func (e *VSphere) DestroyMessage() string {
	return "TODO"
}

// Enterprise return a boolean indicating it is an enterprise provisioner
func (e *VSphere) Enterprise() bool {
	return true
}

const (
	projectPath = "../../../../data/provisioners/cluster/vsphere"
)

func (e VSphere) createVarFile() (err error) {
	var buffer bytes.Buffer
	spec := e.config.Spec.(cfg.VSphere)
	buffer.WriteString(fmt.Sprintf("name = \"%v\"\n", e.config.Metadata.Name))
	buffer.WriteString(fmt.Sprintf("kube_version = \"%v\"\n", spec.Version))
	buffer.WriteString(fmt.Sprintf("env = \"%v\"\n", spec.EnvironmentName))
	buffer.WriteString(fmt.Sprintf("datacenter = \"%v\"\n", spec.Config.DatacenterName))
	// TODO: check plural
	buffer.WriteString(fmt.Sprintf("esxihosts = [\"%v\"]\n", strings.Join(spec.Config.EsxiHost, "\",\"")))
	buffer.WriteString(fmt.Sprintf("datastore = \"%v\"\n", spec.Config.Datastore))
	buffer.WriteString(fmt.Sprintf("network = \"%v\"\n", spec.NetworkConfig.Name))
	buffer.WriteString(fmt.Sprintf("net_cidr = \"%v\"\n", spec.ClusterCIDR))
	buffer.WriteString(fmt.Sprintf("net_gateway = \"%v\"\n", spec.NetworkConfig.Gateway))
	buffer.WriteString(fmt.Sprintf("net_nameservers = [\"%v\"]\n", strings.Join(spec.NetworkConfig.Nameservers, "\",\"")))
	buffer.WriteString(fmt.Sprintf("net_domain = \"%v\"\n", spec.NetworkConfig.Domain))
	buffer.WriteString(fmt.Sprintf("enable_boundary_targets = %v\n", spec.Boundary))
	// TODO: check plural
	buffer.WriteString(fmt.Sprintf("ssh_public_keys = [\"%v\"]\n", strings.Join(spec.SSHPublicKey, "\",\"")))
	buffer.WriteString(fmt.Sprintf("kube_lb_count = %v\n", spec.LoadBalancerNode.Count))
	buffer.WriteString(fmt.Sprintf("kube_lb_template = \"%v\"\n", spec.LoadBalancerNode.Template))
	buffer.WriteString(fmt.Sprintf("kube_lb_custom_script_path = \"%v\"\n", spec.LoadBalancerNode.CustomScriptPath))
	buffer.WriteString(fmt.Sprintf("kube_master_count = %v\n", spec.MasterNode.Count))
	buffer.WriteString(fmt.Sprintf("kube_master_cpu = %v\n", spec.MasterNode.CPU))
	buffer.WriteString(fmt.Sprintf("kube_master_mem = %v\n", spec.MasterNode.MemSize))
	buffer.WriteString(fmt.Sprintf("kube_master_disk_size = %v\n", spec.MasterNode.DiskSize))
	buffer.WriteString(fmt.Sprintf("kube_master_template = \"%v\"\n", spec.MasterNode.Template))
	// TODO: restore
	if len(spec.MasterNode.Labels) > 0 {
		var labels []byte
		labels, err = json.Marshal(spec.MasterNode.Labels)
		if err != nil {
			return err
		}
		buffer.WriteString(fmt.Sprintf("kube_master_labels = %v\n", string(labels)))
	} else {
		buffer.WriteString("kube_master_labels = {}\n")
	}
	if len(spec.MasterNode.Taints) > 0 {
		buffer.WriteString(fmt.Sprintf("kube_master_taints = [\"%v\"]\n", strings.Join(spec.MasterNode.Taints, "\",\"")))
	} else {
		buffer.WriteString("kube_master_taints = []\n")
	}
	// TODO: restore
	// buffer.WriteString(fmt.Sprintf("kube_master_custom_script_path = \"%v\"\n", spec.MasterNode.CustomScriptPath))

	buffer.WriteString(fmt.Sprintf("kube_pod_cidr = \"%v\"\n", spec.ClusterPODCIDR))
	buffer.WriteString(fmt.Sprintf("kube_svc_cidr = \"%v\"\n", spec.ClusterSVCCIDR))

	buffer.WriteString(fmt.Sprintf("kube_infra_count = %v\n", spec.InfraNode.Count))
	buffer.WriteString(fmt.Sprintf("kube_infra_cpu = %v\n", spec.InfraNode.CPU))
	buffer.WriteString(fmt.Sprintf("kube_infra_mem = %v\n", spec.InfraNode.MemSize))
	buffer.WriteString(fmt.Sprintf("kube_infra_disk_size = %v\n", spec.InfraNode.DiskSize))
	buffer.WriteString(fmt.Sprintf("kube_infra_template = \"%v\"\n", spec.InfraNode.Template))
	if len(spec.InfraNode.Labels) > 0 {
		var labels []byte
		labels, err = json.Marshal(spec.InfraNode.Labels)
		if err != nil {
			return err
		}
		buffer.WriteString(fmt.Sprintf("kube_infra_labels = %v\n", string(labels)))
	} else {
		buffer.WriteString("kube_infra_labels = {}\n")
	}
	if len(spec.InfraNode.Taints) > 0 {
		buffer.WriteString(fmt.Sprintf("kube_infra_taints = [\"%v\"]\n", strings.Join(spec.InfraNode.Taints, "\",\"")))
	} else {
		buffer.WriteString("kube_infra_taints = []\n")
	}
	// buffer.WriteString(fmt.Sprintf("kube_infra_custom_script_path = \"%v\"\n", spec.InfraNode..CustomScriptPath))

	if len(spec.NodePools) > 0 {
		buffer.WriteString("node_pools = [\n")
		for _, np := range spec.NodePools {
			buffer.WriteString("{\n")
			buffer.WriteString(fmt.Sprintf("role = \"%v\"\n", np.Role))
			buffer.WriteString(fmt.Sprintf("template = \"%v\"\n", np.Template))
			buffer.WriteString(fmt.Sprintf("count = %v\n", np.Count))
			buffer.WriteString(fmt.Sprintf("memory = %v\n", np.MemSize))
			buffer.WriteString(fmt.Sprintf("cpu = %v\n", np.CPU))
			buffer.WriteString(fmt.Sprintf("disk_size = %v\n", np.DiskSize))
			if len(np.Labels) > 0 {
				var labels []byte
				labels, err = json.Marshal(np.Labels)
				if err != nil {
					return err
				}
				buffer.WriteString(fmt.Sprintf("labels = %v\n", string(labels)))
			} else {
				buffer.WriteString("labels = {}\n")
			}
			if len(np.Taints) > 0 {
				buffer.WriteString(fmt.Sprintf("taints = [\"%v\"]\n", strings.Join(np.Taints, "\",\"")))
			} else {
				buffer.WriteString("taints = []\n")
			}
			// TODO: restore
			buffer.WriteString(fmt.Sprintf("custom_script_path = \"%v\"\n", ""))
			buffer.WriteString("},\n")
		}
		buffer.WriteString("]\n")
	}

	err = ioutil.WriteFile(fmt.Sprintf("%v/vsphere.tfvars", e.terraform.WorkingDir()), buffer.Bytes(), 0600)
	if err != nil {
		return err
	}
	err = e.terraform.FormatWrite(context.Background(), tfexec.Dir(fmt.Sprintf("%v/vsphere.tfvars", e.terraform.WorkingDir())))
	if err != nil {
		return err
	}
	return nil
}

// New instantiates a new GKE provisioner
func New(config *configuration.Configuration) *VSphere {
	b := packr.New("gkecluster", projectPath)
	return &VSphere{
		box:    b,
		config: config,
	}
}

// SetTerraformExecutor adds the terraform executor to this provisioner
func (e *VSphere) SetTerraformExecutor(tf *tfexec.Terraform) {
	e.terraform = tf
}

// TerraformExecutor returns the current terraform executor of this provisioner
func (e *VSphere) TerraformExecutor() (tf *tfexec.Terraform) {
	return e.terraform
}

// Box returns the box that has the files as binary data
func (e VSphere) Box() *packr.Box {
	return e.box
}

// TODO: find Terraform files
// TODO: find Ansible files
// TODO: rename method TerraformFiles() in FilesToBudle()

// TerraformFiles returns the list of files conforming the terraform project
func (e VSphere) TerraformFiles() []string {
	// TODO understand if it is possible to deduce these values somehow
	// find . -type f -follow -print
	return []string{
		"output.tf",
		"main.tf",
		"variables.tf",
		"provision/ansible.cfg",
		"provision/all-in-one.yml",
	}
}

// Prepare the environment before running anything
func (e VSphere) Prepare() error {
	err := createPKI(e.terraform.WorkingDir())
	if err != nil {
		return err
	}
	err = downloadAnsibleRoles(e.terraform.WorkingDir())
	if err != nil {
		return err
	}
	return nil
}

func downloadAnsibleRoles(workingDirectory string) error {
	p_netrc := os.Getenv("NETRC")
	defer os.Setenv("NETRC", p_netrc)

	netrcpath := filepath.Join(workingDirectory, "configuration/.netrc")
	log.Infof("Configuring the NETRC environment variable: %v", netrcpath)
	os.Setenv("NETRC", netrcpath)

	downloadPath := filepath.Join(workingDirectory, "provision/roles")
	log.Infof("Ansible roles download path: %v", downloadPath)
	err := os.Mkdir(downloadPath, fs.FileMode(0755))
	if err != nil {
		return err
	}

	client := &getter.Client{
		Src:  "https://github.com/sighupio/furyctl-provisioners/archive/vsphere.zip//furyctl-provisioners-vsphere/roles",
		Dst:  downloadPath,
		Pwd:  workingDirectory,
		Mode: getter.ClientModeAny,
	}
	err = client.Get()
	if err != nil {
		return err
	}
	return nil
}

// Plan runs a dry run execution
func (e VSphere) Plan() (err error) {
	log.Info("[DRYRUN] Updating VSphere Cluster project")
	// TODO: give the name of the file
	err = e.createVarFile()
	if err != nil {
		return err
	}
	var changes bool
	changes, err = e.terraform.Plan(context.Background(), tfexec.VarFile(fmt.Sprintf("%v/vsphere.tfvars", e.terraform.WorkingDir())))
	if err != nil {
		log.Fatalf("[DRYRUN] Something went wrong while updating gke. %v", err)
		return err
	}
	if changes {
		log.Warn("[DRYRUN] Something changed along the time. Remove dryrun option to apply the desired state")
	} else {
		log.Info("[DRYRUN] Everything is up to date")
	}

	log.Info("[DRYRUN] VSphere Updated")
	return nil
}

// Update runs terraform apply in the project
func (e VSphere) Update() (err error) {
	log.Info("Updating VSphere project")
	err = e.createVarFile()
	if err != nil {
		return err
	}
	err = e.terraform.Apply(context.Background(), tfexec.VarFile(fmt.Sprintf("%v/vsphere.tfvars", e.terraform.WorkingDir())))
	if err != nil {
		log.Fatalf("Something went wrong while updating vsphere. %v", err)
		return err
	}

	log.Info("VSphere Updated")
	return nil
}

// Destroy runs terraform destroy in the project
func (e VSphere) Destroy() (err error) {
	log.Info("Destroying VSphere project")
	err = e.createVarFile()
	if err != nil {
		return err
	}
	err = e.terraform.Destroy(context.Background(), tfexec.VarFile(fmt.Sprintf("%v/vsphere.tfvars", e.terraform.WorkingDir())))
	if err != nil {
		log.Fatalf("Something went wrong while destroying VSphere cluster project. %v", err)
		return err
	}
	log.Info("VSphere destroyed")
	return nil
}

func createPKI(workingDirectory string) error {
	startingPath, err := os.Getwd()
	if err != nil {
		return err
	}

	furyagentPath := filepath.Join(workingDirectory, "furyagent")
	err = os.MkdirAll(furyagentPath, 0755)
	if err != nil {
		return err
	}

	err = os.Chdir(furyagentPath)
	if err != nil {
		return err
	}

	s := storage.Config{
		Provider:  "local",
		LocalPath: ".",
	}

	store, err := storage.Init(&s)
	if err != nil {
		log.Fatal(err)
	}

	var data component.ClusterComponentData = component.ClusterComponentData{
		ClusterConfig: &component.ClusterConfig{},
		Data:          store,
	}

	log.Info("Creating master pki")
	var master component.ClusterComponent = component.Master{
		ClusterComponentData: data,
	}
	err = master.Init("")
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Creating etcd pki")
	var etcd component.ClusterComponent = component.Etcd{
		ClusterComponentData: data,
	}
	err = etcd.Init("")
	if err != nil {
		log.Fatal(err)
	}

	err = os.Chdir(startingPath)
	if err != nil {
		return err
	}

	return nil
}