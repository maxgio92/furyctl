// Copyright (c) 2022 SIGHUP s.r.l All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package configuration

// AWS represents the configuration spec of a AWS bootstrap project including VPC and VPN
type AWS struct {
	NetworkCIDR         string            `yaml:"networkCIDR"`
	PublicSubnetsCIDRs  []string          `yaml:"publicSubnetsCIDRs"`
	PrivateSubnetsCIDRs []string          `yaml:"privateSubnetsCIDRs"`
	VPN                 AWSVPN            `yaml:"vpn"`
	Tags                map[string]string `yaml:"tags"`
}

// AWSVPN represents an VPN configuration
type AWSVPN struct {
	Instances     int      `yaml:"instances"`
	Port          int      `yaml:"port"`
	InstanceType  string   `yaml:"instanceType"`
	DiskSize      int      `yaml:"diskSize"`
	OperatorName  string   `yaml:"operatorName"`
	DHParamsBits  int      `yaml:"dhParamsBits"`
	SubnetCIDR    string   `yaml:"subnetCIDR"`
	SSHUsers      []string `yaml:"sshUsers"`
	OperatorCIDRs []string `yaml:"operatorCIDRs"`
}
