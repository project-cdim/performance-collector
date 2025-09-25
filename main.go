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

package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	logger "github.com/project-cdim/cdim-go-logger"
	logger_common "github.com/project-cdim/cdim-go-logger/common"
	"github.com/project-cdim/performance-collector/controller"
)

// v1 route base url
const URL_BASE_V1 = "/cdim/api/v1"

// Audit Trail Logger
var log, _ = logger.New(logger_common.Option{Tag: logger_common.TAG_TRAIL})

// main is the entry point of the application. It sets up the router with middleware
// and defines the v1 route group with a PUT endpoint for updating Prometheus configurations.
func main() {
	router := gin.Default()
	router.Use(logMiddleware())

	router.Use(cors.New(cors.Config{
		// Allowed Methods
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"},
		// Allowed origins
		AllowOrigins: []string{
			"*",
		},
		// Allowed HTTP request headers
		AllowHeaders: []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
	}))

	// v1 route group
	v1 := router.Group(URL_BASE_V1)

	v1.POST("configs", controller.UpdatePrometheusYaml)

	router.Run(":8080")
}

// logMiddleware returns a gin.HandlerFunc that logs the start and end of each request.
// It logs the HTTP method and URL path at the start of the request and the response status at the end.
func logMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		log.TrailReq(ctx.Request.Method, ctx.Request.URL.Path, "-", "request start.")
		ctx.Next()
		log.TrailRes(ctx.Writer.Status(), "response end.")
	}
}
