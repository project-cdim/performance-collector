// Copyright (C) 2025 NEC Corporation.
// 
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
// 
//     http://www.apache.org/licenses/LICENSE-2.0
// 
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.
        
package model

const (
	DefaultTimeout                     int    = 300
	MaxTimeout                         int    = 3600
	MinTimeout                         int    = 1
	PrometheusYamlFilePath             string = "/files/prometheus.yml"
	PerformanceCollectorConfigFilePath string = "configs/performance-collector.yaml"
	MaxPort                            int    = 65535
	MinPort                            int    = 1
)

type ResourceList struct {
	Count     int `json:"count"`
	Resources []struct {
		Annotation       map[string]any `json:"annotation"`
		Detected         bool           `json:"detected"`
		Device           map[string]any `json:"device"`
		NodeIDs          []string       `json:"nodeIDs"`
		ResourceGroupIDs []string       `json:"resourceGroupIDs"`
	} `json:"resources"`
}

type PrometheusConfig struct {
	PrometheusSettings map[string]any
}

type PerformanceCollectorConfig struct {
	Resources struct {
		TargetUrl string `yaml:"target_url"`
		TimeOut   int    `yaml:"timeout"`
	} `yaml:"resources"`
	Prometheus struct {
		TargetUrl string `yaml:"target_url"`
		TimeOut   int    `yaml:"timeout"`
	} `yaml:"prometheus"`
	Exporter struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"exporter"`
}

// NewPerformanceCollectorConfig creates a new instance of PerformanceCollectorConfig with default timeout values.
// It initializes the Resources and Prometheus TimeOut fields with DefaultTimeout.
func NewPerformanceCollectorConfig() *PerformanceCollectorConfig {
	config := PerformanceCollectorConfig{}
	config.Resources.TimeOut = DefaultTimeout
	config.Prometheus.TimeOut = DefaultTimeout
	return &config
}
