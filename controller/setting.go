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

package controller

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"

	"github.com/project-cdim/performance-collector/internal/model"
	"github.com/project-cdim/performance-collector/internal/service"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

type result struct {
	DeviceID []string `yaml:"deviceID"`
}

var statusToResponse = map[int]gin.H{
	http.StatusInternalServerError: {"code": "0001", "message": "Internal Server Error. Contact the administrator."},
	http.StatusBadRequest:          {"code": "0002", "message": "Bad Request. Check the request parameters."},
	http.StatusNotFound:            {"code": "0003", "message": "Not Found. Check the request URL."},
}

// GetResponseMessage returns a response message based on the provided HTTP status code.
// It looks up the status code in the statusToResponse map and returns the corresponding response.
// If the status code is not found in the map, it panics with an "Unknown status code" message.
func GetResponseMessage(status int) gin.H {
	if response, ok := statusToResponse[status]; ok {
		return response
	}
	panic("Unknown status code: " + strconv.Itoa(status))
}

// logAndRespondWithError logs the error information and sets the response body with an appropriate message.
// It logs the error using the caller's information and the error message, then responds with a JSON message
// based on the provided HTTP status code.
func logAndRespondWithError(c *gin.Context, httpStatus int, caller string, err error) {
	service.Log.Error(caller+" : "+err.Error(), false)
	c.JSON(httpStatus, GetResponseMessage(httpStatus))
}

// UpdatePrometheusYaml updates the Prometheus configuration file with new scrape configurations.
// It reads the existing configuration, merges it with new scrape configurations based on device IDs,
// and writes the updated configuration back to the file.
// The function performs the following steps:
// 1. Reads the existing Prometheus configuration from a file.
// 2. Parses the configuration to extract scrape_configs.
// 3. Sorts the existing scrape_configs by job_name.
// 4. Merges the existing scrape_configs with new configurations based on device IDs.
// 5. Converts the updated configuration to YAML format.
// 6. Writes the updated YAML configuration back to the file.
// If any error occurs during these steps, it logs the error and responds with an appropriate HTTP status code.
func UpdatePrometheusYaml(c *gin.Context) {
	service.Log.Info(c.Request.URL.Path + "[" + c.Request.Method + "] start.")

	// Initialization of performance information collection configuration file
	performanceCollectConfigSettings := model.NewPerformanceCollectorConfig()

	// Load and Validation of the configuration file
	err := loadPerformanceCollector(model.PerformanceCollectorConfigFilePath, performanceCollectConfigSettings)
	if err != nil {
		logAndRespondWithError(c, http.StatusInternalServerError, "UpdatePrometheusYaml() : loadPerformanceCollector", err)
		return
	}
	exporterUrl := performanceCollectConfigSettings.Exporter.Host + ":" + strconv.Itoa(performanceCollectConfigSettings.Exporter.Port)

	// Load of the current prometheus.yml file
	prometheusConfig := model.PrometheusConfig{}
	err = loadPrometheusYaml(model.PrometheusYamlFilePath, &prometheusConfig)
	if err != nil {
		logAndRespondWithError(c, http.StatusInternalServerError, "UpdatePrometheusYaml() : loadPrometheusYaml", err)
		return
	}

	// Retrieve device list: Devices embedded in the node are needed, so they are obtained from the resource list.
	resourcesUrl := performanceCollectConfigSettings.Resources.TargetUrl
	var deviceIds []string
	err = service.GetDeviceList(resourcesUrl, performanceCollectConfigSettings.Resources.TimeOut, &deviceIds)
	if err != nil {
		logAndRespondWithError(c, http.StatusInternalServerError, "UpdatePrometheusYaml() : GetDeviceList", err)
		return
	}
	if len(deviceIds) == 0 {
		service.Log.Warn("UpdatePrometheusYaml() : GetDeviceList : deviceIds is empty")
	}

	// Since the number of devices can be enormous, a sort-merge method is used for faster processing.
	sort.Strings(deviceIds)

	// Retrieve the value of scrape_configs, type assert it, and manipulate it.
	scrapeConfigs := prometheusConfig.PrometheusSettings["scrape_configs"].([]any)

	// Sort scrapeConfigs by job_name
	sort.Slice(scrapeConfigs, func(i, j int) bool {
		jobName1 := scrapeConfigs[i].(map[string]any)["job_name"].(string)
		jobName2 := scrapeConfigs[j].(map[string]any)["job_name"].(string)
		return jobName1 < jobName2
	})

	// Create newScrapeConfigs by sort-merging items from scrapeConfigs that match deviceIds

	newScrapeConfigs := []any{} // Data for newly created ScrapeConfigs

	indexOfScrape := 0
	indexOfDeviceids := 0
	for indexOfScrape < len(scrapeConfigs) && indexOfDeviceids < len(deviceIds) {
		jobName := scrapeConfigs[indexOfScrape].(map[string]any)["job_name"].(string)
		if jobName < deviceIds[indexOfDeviceids] {
			indexOfScrape++
		} else if jobName > deviceIds[indexOfDeviceids] {
			newJob := createNewJob(deviceIds[indexOfDeviceids], exporterUrl)
			newScrapeConfigs = append(newScrapeConfigs, newJob)
			indexOfDeviceids++
		} else {
			// If they match, update targets and add to newScrapeConfigs
			configMap := scrapeConfigs[indexOfScrape].(map[string]any)
			staticConfigs := configMap["static_configs"].([]any)
			// Multiple staticConfigs can be specified, but in this specification, only the first one is updated
			staticConfigMap := staticConfigs[0].(map[string]any)
			targets := staticConfigMap["targets"].([]any)
			// Multiple targets can be specified, but in this specification, only the first one is updated
			targets[0] = exporterUrl
			newScrapeConfigs = append(newScrapeConfigs, scrapeConfigs[indexOfScrape])
			indexOfScrape++
			indexOfDeviceids++
		}
	}
	for indexOfDeviceids < len(deviceIds) {
		newJob := createNewJob(deviceIds[indexOfDeviceids], exporterUrl)
		newScrapeConfigs = append(newScrapeConfigs, newJob)
		indexOfDeviceids++
	}

	// Set the updated data back into the map
	prometheusConfig.PrometheusSettings["scrape_configs"] = newScrapeConfigs

	// Convert the struct to a YAML formatted string
	yamlData, err := yaml.Marshal(&prometheusConfig.PrometheusSettings)
	if err != nil {
		service.Log.Error(err.Error())

		logAndRespondWithError(c, http.StatusInternalServerError, "UpdatePrometheusYaml() : yaml.Marshal", err)
		return
	}

	// Write to file
	err = os.WriteFile(model.PrometheusYamlFilePath, yamlData, fs.FileMode(0o644))
	if err != nil {
		service.Log.Error(err.Error())

		logAndRespondWithError(c, http.StatusInternalServerError, "UpdatePrometheusYaml() : os.WriteFile", err)
		return
	}

	// Request reload to Prometheus
	prometheusReloadUrl := performanceCollectConfigSettings.Prometheus.TargetUrl
	err = service.RequestReload(prometheusReloadUrl, performanceCollectConfigSettings.Prometheus.TimeOut)
	if err != nil {
		logAndRespondWithError(c, http.StatusInternalServerError, "UpdatePrometheusYaml() : RequestReload", err)
		return
	}

	result := result{}
	result.DeviceID = deviceIds

	service.Log.Info(c.Request.URL.Path + "[" + c.Request.Method + "] completed successfully.")

	c.JSON(http.StatusOK, result)
}

// loadPerformanceCollector loads the performance collector settings from the provided configuration.
// It reads the configuration settings related to performance collection, validates them, and initializes
// the performance collector accordingly. This function ensures that all necessary settings are present
// and correctly formatted before initializing the performance collector. If any required setting is missing
// or invalid, it returns an error.
func loadPerformanceCollector(filepath string, settings *model.PerformanceCollectorConfig) error {
	buf, err := os.ReadFile(filepath)
	if err != nil {
		service.Log.Error(err.Error())
		return err
	}

	err = yaml.Unmarshal(buf, &settings)
	if err != nil {
		service.Log.Error(err.Error())
		return err
	}

	if settings.Resources.TargetUrl == "" {
		return errors.New("loadPerformanceCollector : Resources TargetUrl is empty")
	}
	// Only absolute URIs or absolute paths are allowed.
	_, err = url.ParseRequestURI(settings.Resources.TargetUrl)
	if err != nil {
		service.Log.Error(err.Error())
		return err
	}

	resourcesTimeout := settings.Resources.TimeOut
	if resourcesTimeout < model.MinTimeout || resourcesTimeout > model.MaxTimeout {
		return errors.New(fmt.Sprintf("loadPerformanceCollector : Resources Timeout value is out of range. input : %v range : %v - %v", resourcesTimeout, model.MinTimeout, model.MaxTimeout))
	}

	if settings.Prometheus.TargetUrl == "" {
		return errors.New("loadPerformanceCollector : Prometheus TargetUrl is empty.")
	}
	// Only absolute URIs or absolute paths are allowed.
	_, err = url.ParseRequestURI(settings.Prometheus.TargetUrl)
	if err != nil {
		service.Log.Error(err.Error())
		return err
	}

	prometheusTimeout := settings.Prometheus.TimeOut
	if prometheusTimeout < model.MinTimeout || prometheusTimeout > model.MaxTimeout {
		return errors.New(fmt.Sprintf("loadPerformanceCollector : Prometheus Timeout value is out of range. input : %v range : %v - %v", prometheusTimeout, model.MinTimeout, model.MaxTimeout))
	}

	exporterHost := settings.Exporter.Host
	if exporterHost == "" {
		return errors.New("loadPerformanceCollector : Exporter Host value is empty.")
	}

	exporterPort := settings.Exporter.Port
	if exporterPort < model.MinPort || exporterPort > model.MaxPort {
		return errors.New(fmt.Sprintf("loadPerformanceCollector : Exporter Port value is out of range. input : %v range : %v - %v", exporterPort, model.MinPort, model.MaxPort))
	}

	return nil
}

// loadPrometheusYaml loads and validates the Prometheus configuration from a YAML file.
// It checks for the presence of required settings such as global, remote_write, rule_files, and alerting.
// If the scrape_configs setting is missing, it initializes it with an empty array to avoid subsequent reference errors.
// The function returns an error if any of the required settings are missing or if there is an issue with loading the configuration.
func loadPrometheusYaml(filepath string, prometheusConfig *model.PrometheusConfig) error {
	buf, err := os.ReadFile(filepath)
	if err != nil {
		service.Log.Error(err.Error())
		return err
	}

	err = yaml.Unmarshal(buf, &prometheusConfig.PrometheusSettings)
	if err != nil {
		service.Log.Error(err.Error())
		return err
	}

	// Minimum checks
	if !service.ExistsValue(prometheusConfig.PrometheusSettings, "global") {
		return errors.New("loadPrometheusYaml : global setting is required in prometheus.yml")
	}
	if !service.ExistsValue(prometheusConfig.PrometheusSettings, "remote_write") {
		return errors.New("loadPrometheusYaml : remote_write setting is required in prometheus.yml")
	}
	if !service.ExistsValue(prometheusConfig.PrometheusSettings, "rule_files") {
		return errors.New("loadPrometheusYaml : rule_files setting is required in prometheus.yml")
	}
	if !service.ExistsValue(prometheusConfig.PrometheusSettings, "alerting") {
		return errors.New("loadPrometheusYaml : alerting setting is required in prometheus.yml")
	}
	if !service.ExistsValue(prometheusConfig.PrometheusSettings, "scrape_configs") {
		// If scrape_configs does not exist, create an empty array to avoid subsequent reference checks
		prometheusConfig.PrometheusSettings["scrape_configs"] = []any{}
		return nil
	}

	// Type check for job_name and target
	scrapeConfigs, ok := prometheusConfig.PrometheusSettings["scrape_configs"].([]any)
	if !ok {
		return errors.New("loadPrometheusYaml : scrape_configs is not an array in prometheus.yml")
	}
	for _, v := range scrapeConfigs {
		job, ok := v.(map[string]any)
		if !ok {
			return errors.New("loadPrometheusYaml : scrape_config is not a map in prometheus.yml")
		}
		_, ok = job["job_name"].(string)
		if !ok {
			return errors.New("loadPrometheusYaml : job_name is not a string in prometheus.yml")
		}
		staticConfigs, ok := job["static_configs"].([]any)
		if !ok || len(staticConfigs) == 0 {
			return errors.New("loadPrometheusYaml : static_configs is not a slice or is empty in prometheus.yml")
		}
		// Multiple staticConfigs can be specified, but in this specification, only the first one is checked
		staticConfig, ok := staticConfigs[0].(map[string]any)
		if !ok {
			return errors.New("loadPrometheusYaml : static_config is not a map in prometheus.yml")
		}
		targets, ok := staticConfig["targets"].([]any)
		if !ok || len(targets) == 0 {
			return errors.New("loadPrometheusYaml : targets is not a slice or is empty in prometheus.yml")
		}
		// Multiple targets can be specified, but in this specification, only the first one is checked
		_, ok = targets[0].(string)
		if !ok {
			return errors.New("loadPrometheusYaml : target is not a string in prometheus.yml")
		}
	}

	return nil
}

// createNewJob creates a new job configuration for a given device ID and exporter URL.
// It returns a map representing the job configuration, which includes the job name, static configurations with targets,
// and the metrics path specific to the device.
func createNewJob(deviceId string, exporterUrl string) map[string]any {
	return map[string]any{
		"job_name": deviceId,
		"static_configs": []any{
			map[string]any{
				"targets": []any{exporterUrl},
			},
		},
		"metrics_path": "/cdim/api/v1/devices/" + deviceId + "/metrics",
	}
}
