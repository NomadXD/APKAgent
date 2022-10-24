/*
 *  Copyright (c) 2021, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
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

package xds

import (
	"APKAgent/internal/logger"
	"context"
	"io"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/golang/protobuf/ptypes"
	apkmgt_model "github.com/wso2/product-microgateway/adapter/pkg/discovery/api/wso2/discovery/apkmgt"
	stub "github.com/wso2/product-microgateway/adapter/pkg/discovery/api/wso2/discovery/service/apkmgt"

	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	grpcStatus "google.golang.org/grpc/status"
)

var (
	// Last Acknowledged Response from the apkmgt server
	lastAckedResponse *discovery.DiscoveryResponse
	// Last Received Response from the apkmgt server
	// Last Received Response is always is equal to the lastAckedResponse according to current implementation as there is no
	// validation performed on successfully received response.
	lastReceivedResponse *discovery.DiscoveryResponse
	// XDS stream for streaming APIs from APK Mgt client
	xdsStream stub.APKMgtDiscoveryService_StreamAPKMgtApisClient
	//Sent sent the initial
	Sent = false
)

const (
	// The type url for requesting API Entries from apkmgt server.
	apiTypeURL string = "type.googleapis.com/wso2.discovery.apkmgt.Api"
	nodeName          = "mine"
)

// APIEvent represents the event corresponding to a single API Deploy or Remove event
// based on XDS state changes
type APIEvent struct {
	APIUUID      string
	RevisionUUID string
}

func init() {
	lastAckedResponse = &discovery.DiscoveryResponse{}
}

func initConnection(xdsURL string) error {
	// TODO: (AmaliMatharaarachchi) Bring in connection level configurations
	conn, err := grpc.Dial(xdsURL, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		// TODO: (AmaliMatharaarachchi) retries
		logger.LoggerXds.Error("Error while connecting to the APK Management Server.", err)
		return err
	}

	client := stub.NewAPKMgtDiscoveryServiceClient(conn)
	streamContext := context.Background()
	xdsStream, err = client.StreamAPKMgtApis(streamContext)

	if err != nil {
		// TODO: (AmaliMatharaarachchi) handle error.
		logger.LoggerXds.Error("Error while starting client. ", err)
		return err
	}
	logger.LoggerXds.Info("Connection to the APK Management Server: %s is successful.", xdsURL)
	return nil
}

func watchAPIs() {
	for {
		discoveryResponse, err := xdsStream.Recv()
		if err == io.EOF {
			logger.LoggerXds.Error("EOF is received from the apk mgt server.")
			return
		}
		if err != nil {
			logger.LoggerXds.Error("Failed to receive the discovery response ", err)
			errStatus, _ := grpcStatus.FromError(err)
			if errStatus.Code() == codes.Unavailable {
				logger.LoggerXds.Error("Connection stopped. ")
			}
			nack(err.Error())
		} else {
			Sent = true
			lastReceivedResponse = discoveryResponse
			logger.LoggerXds.Debug("Discovery response is received : %s", discoveryResponse.VersionInfo)
			addAPIToChannel(discoveryResponse)
			ack()
		}
	}
}

func ack() {
	lastAckedResponse = lastReceivedResponse
	discoveryRequest := &discovery.DiscoveryRequest{
		Node:          getAdapterNode(),
		VersionInfo:   lastAckedResponse.VersionInfo,
		TypeUrl:       apiTypeURL,
		ResponseNonce: lastReceivedResponse.Nonce,
	}
	xdsStream.Send(discoveryRequest)
}

func nack(errorMessage string) {
	if lastAckedResponse == nil {
		return
	}
	discoveryRequest := &discovery.DiscoveryRequest{
		Node:          getAdapterNode(),
		VersionInfo:   lastAckedResponse.VersionInfo,
		TypeUrl:       apiTypeURL,
		ResponseNonce: lastReceivedResponse.Nonce,
		ErrorDetail: &status.Status{
			Message: errorMessage,
		},
	}
	xdsStream.Send(discoveryRequest)
}

func getAdapterNode() *core.Node {
	return &core.Node{
		Id: nodeName,
	}
}

// InitApkMgtClient initializes the connection to the apkmgt server.
func InitApkMgtClient(xdsURL string) {
	logger.LoggerXds.Info("Starting the XDS Client connection to APK Mgt server.")
	err := initConnection(xdsURL)
	if err == nil {
		go watchAPIs()
		discoveryRequest := &discovery.DiscoveryRequest{
			Node:        getAdapterNode(),
			VersionInfo: "",
			TypeUrl:     apiTypeURL,
		}
		xdsStream.Send(discoveryRequest)
	} else {
		logger.LoggerXds.Error("error in InitApkMgtClient", err.Error())
	}
}

func addAPIToChannel(resp *discovery.DiscoveryResponse) {
	for _, res := range resp.Resources {
		api := &apkmgt_model.Api{}
		err := ptypes.UnmarshalAny(res, api)

		if err != nil {
			logger.LoggerXds.Error("Error while unmarshalling: %s\n", err.Error())
			continue
		}
		logger.LoggerXds.Debug("client has received: ", res)
	}
}
