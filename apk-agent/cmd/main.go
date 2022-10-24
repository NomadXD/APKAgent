/*
 *  Copyright (c) 2022, WSO2 Inc. (http://www.wso2.org)
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 */

package main

import (
	"os"
	"os/signal"

	"github.com/AmaliMatharaarachchi/APKAgent/apk-agent/internal/logger"
	"github.com/AmaliMatharaarachchi/APKAgent/apk-agent/internal/operator"
	"github.com/AmaliMatharaarachchi/APKAgent/apk-agent/internal/xds"
)

const (
	maxRandomInt             int = 999999999
	grpcMaxConcurrentStreams     = 1000000
	address                      = "localhost:18000"
)

func main() {
	logger.LoggerServer.Info("Hello, world from Agent.")
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)
	// Start APK management client
	go xds.InitApkMgtClient(address)
	// Start APK agent operator
	go operator.Init()

OUTER:
	for {
		select {
		case s := <-sig:
			switch s {
			case os.Interrupt:
				break OUTER
			}
		}
	}
}
