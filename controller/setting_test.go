// Copyright (C) 2025-2026 NEC Corporation.
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
        
package controller

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/project-cdim/performance-collector/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestUpdatePrometheusYaml(t *testing.T) {
	t.Skip("Skip: TestUpdatePrometheusYaml")
}

func Test_loadPerformanceCollector(t *testing.T) {
	const settingsPath = "testdata/"
	yamlContent := model.PerformanceCollectorConfig{}
	type args struct {
		filepath string
		settings *model.PerformanceCollectorConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"normal", args{settingsPath + "collector_normal.yaml", &yamlContent}, false},
		{"not found", args{settingsPath + "collector_notFound.yaml", &yamlContent}, true},
		{"unmarshal", args{settingsPath + "collector_unmarshal.yaml", &yamlContent}, true},
		{"resourcesTargetUrlIsNull", args{settingsPath + "collector_resourcesTargetUrlIsNull.yaml", &yamlContent}, true},
		{"resourcesTargetUrl", args{settingsPath + "collector_resourcesTargetUrl.yaml", &yamlContent}, true},
		{"resourcesTimeout1", args{settingsPath + "collector_resourcesTimeout1.yaml", &yamlContent}, true},
		{"resourcesTimeout2", args{settingsPath + "collector_resourcesTimeout2.yaml", &yamlContent}, true},
		{"prometheusTargetUrliIsNull", args{settingsPath + "collector_prometheusTargetUrlIsNull.yaml", &yamlContent}, true},
		{"prometheusTargetUrl", args{settingsPath + "collector_prometheusTargetUrl.yaml", &yamlContent}, true},
		{"prometheusTimeout1", args{settingsPath + "collector_prometheusTimeout1.yaml", &yamlContent}, true},
		{"prometheusTimeout2", args{settingsPath + "collector_prometheusTimeout2.yaml", &yamlContent}, true},
		{"exporterHost", args{settingsPath + "collector_exporterHost.yaml", &yamlContent}, true},
		{"exporterPort1", args{settingsPath + "collector_exporterPort1.yaml", &yamlContent}, true},
		{"exporterPort2", args{settingsPath + "collector_exporterPort2.yaml", &yamlContent}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.settings = &model.PerformanceCollectorConfig{}
			if err := loadPerformanceCollector(context.Background(), tt.args.filepath, tt.args.settings); (err != nil) != tt.wantErr {
				t.Errorf("loadConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_loadPrometheusYaml(t *testing.T) {
	const settingsPath = "testdata/"
	yamlContent := model.PrometheusConfig{}
	type args struct {
		filepath         string
		prometheusConfig *model.PrometheusConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"normal", args{settingsPath + "prometheus_normal.yaml", &yamlContent}, false},
		{"prometheus_error_31", args{settingsPath + "prometheus_notFound.yaml", &yamlContent}, true},
		{"prometheus_error_32", args{settingsPath + "prometheus_unmarshal.yaml", &yamlContent}, true},
		{"prometheus_error_33", args{settingsPath + "prometheus_global.yaml", &yamlContent}, true},
		{"prometheus_error_34", args{settingsPath + "prometheus_remote.yaml", &yamlContent}, true},
		{"prometheus_error_35", args{settingsPath + "prometheus_rule.yaml", &yamlContent}, true},
		{"prometheus_error_36", args{settingsPath + "prometheus_alerting.yaml", &yamlContent}, true},
		{"prometheus_error_37", args{settingsPath + "prometheus_scrape_config37.yaml", &yamlContent}, false},
		{"prometheus_error_38", args{settingsPath + "prometheus_scrape_config38.yaml", &yamlContent}, true},
		{"prometheus_error_39", args{settingsPath + "prometheus_jobname39.yaml", &yamlContent}, true},
		{"prometheus_error_40", args{settingsPath + "prometheus_error40.yaml", &yamlContent}, true},
		{"prometheus_error_41", args{settingsPath + "prometheus_error41.yaml", &yamlContent}, true},
		{"prometheus_error_42", args{settingsPath + "prometheus_error42.yaml", &yamlContent}, true},
		{"prometheus_error_43", args{settingsPath + "prometheus_error43.yaml", &yamlContent}, true},
		{"prometheus_error_44", args{settingsPath + "prometheus_scrape_config44.yaml", &yamlContent}, true},
	}
	for _, tt := range tests {
		tt.args.prometheusConfig = &model.PrometheusConfig{}
		t.Run(tt.name, func(t *testing.T) {
			if err := loadPrometheusYaml(context.Background(), tt.args.filepath, tt.args.prometheusConfig); (err != nil) != tt.wantErr {
				t.Errorf("loadPrometheusSettingFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetResponseMessage(t *testing.T) {
	tests := []struct {
		status   int
		expected gin.H
	}{
		{http.StatusInternalServerError, gin.H{"code": "0001", "message": "Internal Server Error. Contact the administrator."}},
		{http.StatusBadRequest, gin.H{"code": "0002", "message": "Bad Request. Check the request parameters."}},
		{http.StatusNotFound, gin.H{"code": "0003", "message": "Not Found. Check the request URL."}},
	}

	for _, tt := range tests {
		t.Run(tt.expected["message"].(string), func(t *testing.T) {
			response := GetResponseMessage(tt.status)
			assert.Equal(t, tt.expected, response)
		})
	}
}

func Test_logAndRespondWithError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode int
		expectedBody string
	}{
		{
			name:         "BadRequestError",
			err:          fmt.Errorf(`{"code": "0002", "message": "Bad Request. Check the request parameters."}`),
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code": "0002", "message": "Bad Request. Check the request parameters."}`,
		},
		{
			name:         "InternalServerError",
			err:          fmt.Errorf(`{"code":"0001","message":"An internal error occurred"}`),
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"code":"0001", "message":"Internal Server Error. Contact the administrator."}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup Gin context and response recorder
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

			// Execute the function
			logAndRespondWithError(c.Request.Context(), c, tc.expectedCode, "TestError", tc.err)

			// Verify the status code and response body
			assert.Equal(t, tc.expectedCode, w.Code)
			assert.JSONEq(t, tc.expectedBody, w.Body.String())
		})
	}
}

func Test_createNewJob(t *testing.T) {
	type args struct {
		deviceId    string
		exporterUrl string
	}
	tests := []struct {
		name string
		args args
		want map[string]any
	}{
		{
			name: "SuccessCase",
			args: args{
				deviceId:    "device123",
				exporterUrl: "http://example.com/exporter",
			},
			want: map[string]any{
				"job_name":     "device123",
				"metrics_path": "/cdim/api/v1/devices/device123/metrics",
				"static_configs": []map[string]any{
					{
						"targets": []string{"http://example.com/exporter"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := createNewJob(tt.args.deviceId, tt.args.exporterUrl)
			if got["job_name"] != tt.want["job_name"] {
				t.Errorf("createNewJob() = %v, want %v", got["job_name"], tt.want["job_name"])
			}
			if got["metrics_path"] != tt.want["metrics_path"] {
				t.Errorf("createNewJob() = %v, want %v", got["metrics_path"], tt.want["metrics_path"])
			}
			if got["targets"] != tt.want["targets"] {
				t.Errorf("createNewJob() = %v, want %v", got["targets"], tt.want["targets"])
			}
		})
	}
}
