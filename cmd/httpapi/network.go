/*
Copyright 2024 DCN

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

package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	// Meta API for object metadata

	v1beta1 "dcnlab.ssu.ac.kr/kt-cloud-operator/api/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

type PublicNetwork struct {
	NcListentPublicIpsResponse NcListentPublicIpsResponse `json:"nc_listentpublicipsresponse"`
}

type NcListentPublicIpsResponse struct {
	PublicIps []PublicIp `json:"publicips"`
}

type PublicIp struct {
	EntPublicCIDRId string      `json:"entpubliccidrid"`
	VirtualIps      []VirtualIp `json:"virtualips"`
	VPCId           string      `json:"vpcid"`
	IP              string      `json:"ip"`
	ZoneId          string      `json:"zoneid"`
	Id              string      `json:"id"`
	Type            string      `json:"type"`
	Account         string      `json:"account"`
}

type VirtualIp struct {
	VMGuestIP   string `json:"vmguestip"`
	IPAddress   string `json:"ipaddress"`
	VPCId       string `json:"vpcid"`
	IPAddressId string `json:"ipaddressid"`
	Name        string `json:"name"`
	NetworkId   string `json:"networkid"`
	Id          string `json:"id"`
}

// Post Request Payload attach nat
type PostPayload struct {
	VMGuestIP     string `json:"vmguestip"`
	VMNetworkId   string `json:"vmnetworkid"`
	EntPublicIPId string `json:"entpublicipid"`
}

// attach NAT response
type NATAttachResponse struct {
	NcEnableStaticNatResponse NcEnableStaticNatResponse `json:"nc_enablestaticnatresponse"`
}

type NcEnableStaticNatResponse struct {
	DisplayText string `json:"displaytext"`
	Success     bool   `json:"success"`
}

func AttachPublicIP(machine *v1beta1.KTMachine, token string) error {

	var machinePrivateAddresses []string

	// Iterate over dynamic keys in "addresses"
	for network, addresses := range machine.Status.Addresses {
		fmt.Printf("Network: %s\n", network)
		for _, addr := range addresses {
			fmt.Printf("  Address: %s\n", addr.Addr)
			fmt.Printf("  Version: %d\n", addr.Version)
			machinePrivateAddresses = append(machinePrivateAddresses, addr.Addr)
		}
	}

	if len(machinePrivateAddresses) == 0 {
		return errors.New("failed to get machine address to pair with public ip address for snat")
	}

	vmguestip := machinePrivateAddresses[0]       //just get the first IP address
	vmnetworkid := machine.Spec.NetworkTier[0].ID //just get the first tier

	publicIPs, err := GetAvailablePublicIpAddresses(token)

	// publicIPsJson, err := json.Marshal(publicIPs)
	// if err != nil {
	// 	logger1.Error("Error marshalling publicIPs to JSON:", err)
	// 	return err
	// }
	// logger1.Info("Response Body Networks GOT Filtered:", string(publicIPsJson))

	// logger1.Info("==================================================================")
	// fmt.Println(string(machinePrivateAddresses[0]))
	// logger1.Info("------------------------------")

	if err != nil {
		return err
	}
	if len(publicIPs.PublicIps) == 0 {
		return errors.New("no available public ip addresses on the cloud, maybe try creating in the cloud in same zone as the cluster")
	}
	entpublicipid := publicIPs.PublicIps[0].Id

	networkAttachRequest := PostPayload{
		VMGuestIP:     vmguestip,
		VMNetworkId:   vmnetworkid,
		EntPublicIPId: entpublicipid,
	}

	// Marshal the struct to JSON
	payload, err := json.Marshal(networkAttachRequest)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return err
	}

	// Define the endpoint URL
	apiURL := Config.ApiBaseURL + Config.Zone + "/nc/StaticNat"

	// Set up HTTP client with timeout
	// Set up the HTTP client
	client := &http.Client{Timeout: 10 * time.Second}

	// Create a new HTTP POST request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payload))
	if err != nil {
		logger1.Error("Error creating request:", err)
		return err
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Auth-Token", token) // Replace with your actual token

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		logger1.Error("Error sending request:", err)
		return err
	}
	defer resp.Body.Close()

	// Handle the response
	fmt.Println("Response Status:", resp.Status)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger1.Error("Error reading response body:", err)
		return err
	}
	logger1.Info("Response Body:", string(body))

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		logger1.Info("POST request successful and attached Ip Address to machine!")

		// Parse the JSON into the struct
		var serverResponse NATAttachResponse
		err = json.Unmarshal(body, &serverResponse)
		if err != nil {
			logger1.Error("Error unmarshaling JSON response:", err)
			return err
		}

		// logger1.Info("Response Text: " + serverResponse.NcEnableStaticNatResponse.DisplayText)

		if !serverResponse.NcEnableStaticNatResponse.Success {
			return errors.New(serverResponse.NcEnableStaticNatResponse.DisplayText)
		}

		// logger1.Info("Didnt pass here")

		// Update the machine
		// Update the machine K8s Resource
		clientConfig, err := getRestConfig(Config.Kubeconfig)
		if err != nil {
			logger1.Errorf("Cannot prepare k8s client config: %v. Kubeconfig was: %s", err, Config.Kubeconfig)
			return err
		}
		// Set up a scheme (use runtime.Scheme from apimachinery)
		scheme := runtime.NewScheme()
		// Create Kubernetes client
		k8sClient, err := getClient(clientConfig, scheme)
		if err != nil {
			logger1.Fatalf("Failed to create Kubernetes client: %v", err)
			return err
		}
		machineStatusCopy := machine.Status
		assignedIp := v1beta1.AssignedPublicIps{
			Id: publicIPs.PublicIps[0].Id,
			IP: publicIPs.PublicIps[0].IP,
		}
		machineStatusCopy.AssignedPublicIps = append(machineStatusCopy.AssignedPublicIps, assignedIp)

		err = updateVMStatus(k8sClient, machine, &machineStatusCopy, machineStatusCopy.Status)
		if err != nil {
			logger1.Errorf("Failed to update VMstatus with public IP: %v", err)
			return err
		}
		logger1.Info("Updated the status of machine with public IP")
		return nil

	} else {
		logger1.Error("POST request failed with status:", resp.Status)
		return errors.New("post request failed with status:" + resp.Status)
	}
}

// get all public IPs
func GetAvailablePublicIpAddresses(token string) (NcListentPublicIpsResponse, error) {

	// Define the API URL
	apiURL := Config.ApiBaseURL + Config.Zone + "/nc/IpAddress"

	// Set up the HTTP client
	client := &http.Client{Timeout: 10 * time.Second}

	// Create a new HTTP GET request
	req, err := http.NewRequest("GET", apiURL, bytes.NewBuffer([]byte{}))
	if err != nil {
		logger1.Error("Error creating GET VM request:", err)
		return NcListentPublicIpsResponse{}, err
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Auth-Token", token) // Replace with actual token

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		logger1.Error("Error sending request:", err)
		return NcListentPublicIpsResponse{}, err
	}
	defer resp.Body.Close()

	// Handle the response
	fmt.Println("Response Status:", resp.Status)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger1.Error("Error reading response body:", err)
		return NcListentPublicIpsResponse{}, err
	}

	// logger1.Info("-----------------------------------------")
	// logger1.Info("Response Body Networks:", string(body))
	// logger1.Info("********************************")

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		logger1.Info("GET request successful and got machine!")
		// Parse the JSON into the struct
		var serverResponse PublicNetwork
		err = json.Unmarshal(body, &serverResponse)
		if err != nil {
			logger1.Error("Error unmarshaling JSON response:", err)
			return NcListentPublicIpsResponse{}, err
		}

		filteredResponse := NcListentPublicIpsResponse{}
		filteredPublicIps := []PublicIp{}
		for i := 0; i < len(serverResponse.NcListentPublicIpsResponse.PublicIps); i++ {
			publicIps := serverResponse.NcListentPublicIpsResponse.PublicIps
			if len(publicIps[i].VirtualIps) == 0 && publicIps[i].Type == "ASSOCIATE" {
				publicIP := PublicIp{
					EntPublicCIDRId: publicIps[i].EntPublicCIDRId,
					VirtualIps:      publicIps[i].VirtualIps,
					VPCId:           publicIps[i].VPCId,
					IP:              publicIps[i].IP,
					ZoneId:          publicIps[i].ZoneId,
					Type:            publicIps[i].Type,
					Id:              publicIps[i].Id,
					Account:         publicIps[i].Account,
				}
				filteredPublicIps = append(filteredPublicIps, publicIP)
			}
		}
		filteredResponse.PublicIps = filteredPublicIps

		return filteredResponse, nil

	} else {
		logger1.Error("GET request failed with status:", resp.Status)
		return NcListentPublicIpsResponse{}, errors.New("GET request failed with status: " + resp.Status)
	}

}
