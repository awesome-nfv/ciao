//
// Copyright (c) 2016 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package configuration

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/01org/ciao/payloads"
	"gopkg.in/yaml.v2"
)

// we can have values set to default, except for
//    scheduler { storage_uri }
//    controller { compute_ca, compute_cert, identity_user, identity_password }
//    launcher { compute_net, mgmt_net }
//    image_service { url }
//    identity_service { url }
//
// so we need to have at least those values set in our config
//
// TODO: proper validation of values set in yaml setup
func validMinConf(conf *payloads.Configure) bool {
	return (conf.Configure.Scheduler.ConfigStorageURI != "" &&
		conf.Configure.Controller.HTTPSCACert != "" &&
		conf.Configure.Controller.HTTPSKey != "" &&
		conf.Configure.Controller.IdentityUser != "" &&
		conf.Configure.Controller.IdentityPassword != "" &&
		conf.Configure.Launcher.ComputeNetwork != "" &&
		conf.Configure.Launcher.ManagementNetwork != "" &&
		conf.Configure.ImageService.URL != "" &&
		conf.Configure.IdentityService.URL != "")
}

func fillDefaults(conf *payloads.Configure) {
	conf.Configure.Scheduler.ConfigStorageType = payloads.Filesystem
	conf.Configure.Controller.ComputePort = 8774
	conf.Configure.ImageService.Type = payloads.Glance
	conf.Configure.IdentityService.Type = payloads.Keystone
	conf.Configure.Launcher.DiskLimit = true
	conf.Configure.Launcher.MemoryLimit = true
}

// TODO: add etcd support related scheme(s)
func discoverDriver(uriStr string) (storageType payloads.StorageType, err error) {
	uri, err := url.Parse(uriStr)
	if err != nil {
		return storageType, err
	}
	switch uri.Scheme {
	case "file":
		return payloads.Filesystem, nil
	default:
		return "", fmt.Errorf(
			"Configuration URI Scheme '%s' not supported", uri.Scheme)
	}
}

// Payload fills the payloads.Configure struct passed in 'conf'
// with the values from the bytes given
func Payload(yamlConf []byte, conf *payloads.Configure) (err error) {
	if yamlConf == nil {
		return fmt.Errorf("Unable to retrieve configuration from empty definition")
	}
	err = yaml.Unmarshal(yamlConf, &conf)
	if err != nil {
		return err
	}
	return nil
}

// Blob returns an array of bytes containing
// the cluster configuration.
func Blob(conf *payloads.Configure) (blob []byte, err error) {
	if validMinConf(conf) == false {
		return nil, errors.New(
			"minimal configuration is not met or yaml is malformed")
	}
	blob, err = yaml.Marshal(&conf)
	if err != nil {
		return nil, err
	}
	return blob, nil
}

// ExtractBlob returns a configuration payload.
// It could be used by the SSNTP server or some other entity.
func ExtractBlob(uri string) (payload []byte, err error) {
	var d driver
	driverType, err := discoverDriver(uri)
	if err != nil {
		return nil, err
	}
	switch driverType {
	case payloads.Filesystem:
		d = &file{}
	case payloads.Etcd:
		d = &etcd{}
	}
	conf, err := d.fetchConfiguration(uri)
	if err != nil {
		return nil, err
	}
	payload, err = Blob(&conf)
	if err != nil {
		return nil, err
	}
	return payload, nil
}
