/*
Copyright 2020 The Kubernetes Authors.

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

package utils

import (
	"flag"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

var (
	configContent = flag.String("nfd.e2e-config", "", "Configuration parameters for end-to-end tests")
	E2EConfigFile *e2eConfig
)

type e2eConfig struct {
	DefaultFeatures *struct {
		LabelWhitelist      lookupMap
		AnnotationWhitelist lookupMap
		Nodes               map[string]nodeConfig
	}
}

type nodeConfig struct {
	ExpectedLabelValues      map[string]string
	ExpectedLabelKeys        lookupMap
	ExpectedAnnotationValues map[string]string
	ExpectedAnnotationKeys   lookupMap
}

type lookupMap map[string]struct{}

func (l *lookupMap) UnmarshalJSON(data []byte) error {
	*l = lookupMap{}

	var slice []string
	if err := yaml.Unmarshal(data, &slice); err != nil {
		return err
	}

	for _, k := range slice {
		(*l)[k] = struct{}{}
	}
	return nil
}

func ReadConfig() error {
	// Read and parse only once
	if E2EConfigFile != nil || *configContent == "" {
		return nil
	}

	data, err := ioutil.ReadFile(*configContent)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, E2EConfigFile); err != nil {
		return err
	}

	return nil
}
