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
        
package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/project-cdim/performance-collector/internal/model"
)

func TestGetDeviceList(t *testing.T) {
	// Create a test server to simulate the external web application
	testServerOk := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	}))
	defer testServerOk.Close()

	testServerNg := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "ERROR"}`))
	}))
	defer testServerNg.Close()

	type args struct {
		resourcesUrl string
		deviceIds    *[]string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid URL with device IDs",
			args: args{
				resourcesUrl: testServerOk.URL,
				deviceIds:    &[]string{},
			},
			wantErr: false,
		},
		{
			name: "invalid URL",
			args: args{
				resourcesUrl: testServerNg.URL,
				deviceIds:    &[]string{},
			},
			wantErr: true,
		},
		{
			name: "empty URL",
			args: args{
				resourcesUrl: "",
				deviceIds:    &[]string{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := GetDeviceList(tt.args.resourcesUrl, 300, tt.args.deviceIds); (err != nil) != tt.wantErr {
				t.Errorf("GetDeviceList() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExistsValue(t *testing.T) {
	type args struct {
		target map[string]any
		item   string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "value exists in map",
			args: args{
				target: map[string]any{"key1": "value1", "key2": "value2"},
				item:   "key1",
			},
			want: true,
		},
		{
			name: "value does not exist in map",
			args: args{
				target: map[string]any{"key1": "value1", "key2": "value2"},
				item:   "key3",
			},
			want: false,
		},
		{
			name: "empty map",
			args: args{
				target: map[string]any{},
				item:   "key1",
			},
			want: false,
		},
		{
			name: "nil map",
			args: args{
				target: nil,
				item:   "key1",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExistsValue(tt.args.target, tt.args.item); got != tt.want {
				t.Errorf("IsExistValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequestReload(t *testing.T) {
	// Create a test server to simulate the external web application
	testServerOk := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	}))
	defer testServerOk.Close()

	testServerNg := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "ERROR"}`))
	}))
	defer testServerNg.Close()

	type args struct {
		prometheusReloadUrl string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid URL without error",
			args: args{
				prometheusReloadUrl: testServerOk.URL,
			},
			wantErr: false,
		},
		{
			name: "invalid URL with error",
			args: args{
				prometheusReloadUrl: testServerNg.URL,
			},
			wantErr: true,
		},
		{
			name: "empty URL with error",
			args: args{
				prometheusReloadUrl: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := RequestReload(tt.args.prometheusReloadUrl, 3000); (err != nil) != tt.wantErr {
				t.Errorf("RequestReload() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_extractDeviceIDs(t *testing.T) {
	type args struct {
		resourceList model.ResourceList
		deviceIds    *[]string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "empty response",
			args: args{
				resourceList: model.ResourceList{},
				deviceIds:    &[]string{},
			},
			wantErr: false,
		},
		{
			name: "valid response with single device",
			args: args{
				resourceList: model.ResourceList{
					Count: 1,
					Resources: []struct {
						Annotation       map[string]any `json:"annotation"`
						Detected         bool           `json:"detected"`
						Device           map[string]any `json:"device"`
						NodeIDs          []string       `json:"nodeIDs"`
						ResourceGroupIDs []string       `json:"resourceGroupIDs"`
					}{
						{
							Detected: true,
							Device: map[string]any{
								"deviceID": "testDevice",
							},
							NodeIDs: []string{"node1"},
						},
					},
				},
				deviceIds: &[]string{},
			},
			wantErr: false,
		},
		{
			name: "valid response with multiple devices",
			args: args{
				resourceList: model.ResourceList{
					Count: 2,
					Resources: []struct {
						Annotation       map[string]any `json:"annotation"`
						Detected         bool           `json:"detected"`
						Device           map[string]any `json:"device"`
						NodeIDs          []string       `json:"nodeIDs"`
						ResourceGroupIDs []string       `json:"resourceGroupIDs"`
					}{
						{
							Detected: true,
							Device: map[string]any{
								"deviceID": "device123",
							},
							NodeIDs: []string{"node1"},
						},
						{
							Detected: true,
							Device: map[string]any{
								"deviceID": "device456",
							},
							NodeIDs: []string{"node1"},
						},
					},
				},
				deviceIds: &[]string{},
			},
			wantErr: false,
		},
		{
			name: "invalid response with single device",
			args: args{
				resourceList: model.ResourceList{
					Count: 1,
					Resources: []struct {
						Annotation       map[string]any `json:"annotation"`
						Detected         bool           `json:"detected"`
						Device           map[string]any `json:"device"`
						NodeIDs          []string       `json:"nodeIDs"`
						ResourceGroupIDs []string       `json:"resourceGroupIDs"`
					}{
						{
							Detected: false,
							Device: map[string]any{
								"deviceID": "testDevice",
							},
							NodeIDs: []string{"node1"},
						},
					},
				},
				deviceIds: &[]string{},
			},
			wantErr: false,
		},
		{
			name: "Error response with single device",
			args: args{
				resourceList: model.ResourceList{
					Count: 1,
					Resources: []struct {
						Annotation       map[string]any `json:"annotation"`
						Detected         bool           `json:"detected"`
						Device           map[string]any `json:"device"`
						NodeIDs          []string       `json:"nodeIDs"`
						ResourceGroupIDs []string       `json:"resourceGroupIDs"`
					}{
						{
							Detected: true,
							Device: map[string]any{
								"deviceID": true,
							},
							NodeIDs: []string{"node1"},
						},
					},
				},
				deviceIds: &[]string{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := extractDeviceIDs(tt.args.resourceList, tt.args.deviceIds); (err != nil) != tt.wantErr {
				t.Errorf("extractDeviceIDs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
