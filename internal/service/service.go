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
        
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/project-cdim/performance-collector/internal/model"
	loggercommon "github.com/project-cdim/cdim-go-logger/common"
)

// GetDeviceList retrieves a list of devices from the specified resources URL.
// It uses an HTTP client with a specified timeout to make the request.
// The device IDs are extracted from the response and stored in the provided deviceIds slice
//
// Parameters:
//   - ctx: context.Context. The context for the HTTP request and logging.
//   - resourcesUrl: string. The URL to fetch the device list from.
//   - timeout: int. The timeout in seconds for the HTTP request.
//   - deviceIds: *[]string. The slice to be populated with device IDs.
//
// Returns:
//   - error: Returns an error if the HTTP request fails or response cannot be processed, nil otherwise.
func GetDeviceList(ctx context.Context, resourcesUrl string, timeout int, deviceIds *[]string) error {
	httpClient := http.Client{Timeout: time.Duration(timeout) * time.Second}

	// Create HTTP request with context
	req, newReqErr := http.NewRequestWithContext(ctx, http.MethodGet, resourcesUrl, nil)
	if newReqErr != nil {
		Log.ErrorWithContext(ctx, newReqErr.Error())
		return newReqErr
	}

	// Add traceparent header if present in context
	traceparent := loggercommon.TraceparentFromContext(ctx)
	if traceparent != "" {
		req.Header.Set("traceparent", traceparent)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		Log.ErrorWithContext(ctx, err.Error())
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("GetDeviceList : Failed to get resources. resp.StatusCode:%v", resp.StatusCode))
	}

	// Read the response body
	body, err := io.ReadAll(io.Reader(resp.Body))
	if err != nil {
		Log.ErrorWithContext(ctx, err.Error())
		return err
	}
	// Parse the response
	response := model.ResourceList{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		Log.ErrorWithContext(ctx, err.Error())
		return err
	}

	err = extractDeviceIDs(response, deviceIds)
	if err != nil {
		return err
	}

	return nil
}

// extractDeviceIDs extracts device IDs from the given resource list and appends them to the provided deviceIds slice.
// It iterates over the resources in the resource list, and for each resource that is detected and has node IDs,
// it checks if the device ID is a string and appends it to the deviceIds slice.
// If the device ID is not a string, it returns an error
func extractDeviceIDs(resourceList model.ResourceList, deviceIds *[]string) error {
	for _, r := range resourceList.Resources {
		if r.Detected && len(r.NodeIDs) > 0 {
			if deviceID, ok := r.Device["deviceID"].(string); ok {
				*deviceIds = append(*deviceIds, deviceID)
			} else {
				return errors.New("extractDeviceIDs : deviceID is not string.")
			}
		}
	}
	return nil
}

// ExistsValue checks if a given item exists in the target map and is not nil.
// It returns true if the item exists and is not nil, otherwise it returns false
func ExistsValue(target map[string]any, item string) bool {
	if targetValue, ok := target[item]; ok && targetValue != nil {
		return true
	}
	return false
}

// RequestReload sends a POST request to the specified Prometheus reload URL to trigger a configuration reload.
// It uses an HTTP client with a specified timeout to make the request.
// If the request fails or the response status is not OK, it logs the error and returns it
//
// Parameters:
//   - ctx: context.Context. The context for the HTTP request and logging.
//   - prometheusReloadUrl: string. The URL to send the reload request to.
//   - timeout: int. The timeout in seconds for the HTTP request.
//
// Returns:
//   - error: Returns an error if the HTTP request fails or response status is not OK, nil otherwise.
func RequestReload(ctx context.Context, prometheusReloadUrl string, timeout int) error {
	httpClient := http.Client{Timeout: time.Duration(timeout) * time.Second}

	// Create HTTP request with context
	req, newReqErr := http.NewRequestWithContext(ctx, http.MethodPost, prometheusReloadUrl, nil)
	if newReqErr != nil {
		Log.ErrorWithContext(ctx, newReqErr.Error())
		return newReqErr
	}

	// Add traceparent header if present in context
	traceparent := loggercommon.TraceparentFromContext(ctx)
	if traceparent != "" {
		req.Header.Set("traceparent", traceparent)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		Log.ErrorWithContext(ctx, err.Error())
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("RequestReload : Failed to send reload request to Prometheus. resp.StatusCode:%v", resp.StatusCode))
	}

	return nil
}
