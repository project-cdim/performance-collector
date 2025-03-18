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

import (
	"reflect"
	"testing"
)

func TestNewPerformanceCollectorConfig(t *testing.T) {
	tests := []struct {
		name string
		want *PerformanceCollectorConfig
	}{
		{
			name: "DefaultConfig",
			want: &PerformanceCollectorConfig{
				Resources: struct {
					TargetUrl string "yaml:\"target_url\""
					TimeOut   int    "yaml:\"timeout\""
				}{TargetUrl: "", TimeOut: 300},
				Prometheus: struct {
					TargetUrl string "yaml:\"target_url\""
					TimeOut   int    "yaml:\"timeout\""
				}{TargetUrl: "", TimeOut: 300},
				Exporter: struct {
					Host string "yaml:\"host\""
					Port int    "yaml:\"port\""
				}{Host: "", Port: 0},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewPerformanceCollectorConfig(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPerformanceCollectorConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
