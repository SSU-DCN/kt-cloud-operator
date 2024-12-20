package httpapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"dcnlab.ssu.ac.kr/kt-cloud-operator/api/v1beta1"
)

// Meta API for object metadata

type KTMachineStatus struct {
	ID             string           `json:"id"`
	AdminPass      string           `json:"adminPass"`
	Links          []Links          `json:"links"`
	SecurityGroups []SecurityGroups `json:"securityGroups"`
}

type Links struct {
	Rel  string `json:"rel,omitempty"`
	Href string `json:"href,omitempty"`
}

type SecurityGroups struct {
	Name string `json:"name,omitempty"`
}

// For posting to create machine
type Network struct {
	UUID string `json:"uuid"`
}

type BlockDeviceMappingV2 struct {
	DestinationType string `json:"destination_type"`
	BootIndex       int    `json:"boot_index"`
	SourceType      string `json:"source_type"`
	VolumeSize      int    `json:"volume_size"`
	UUID            string `json:"uuid"`
}

type Server struct {
	Name                 string                 `json:"name"`
	KeyName              string                 `json:"key_name"`
	FlavorRef            string                 `json:"flavorRef"`
	AvailabilityZone     string                 `json:"availability_zone"`
	Networks             []Network              `json:"networks"`
	BlockDeviceMappingV2 []BlockDeviceMappingV2 `json:"block_device_mapping_v2"`
}

type RequestPayload struct {
	Server Server `json:"server"`
}

func CreateVM(machine *v1beta1.KTMachine, token string) error {
	// Create the payload
	networks := []Network{}
	block_device_mapping_v2 := []BlockDeviceMappingV2{}

	for i, network := range machine.Spec.Networks {
		fmt.Println(network.ID, i)
		networks = append(
			networks,
			Network{
				UUID: network.ID,
			})
	}

	for i, block_device_mapping := range machine.Spec.BlockDeviceMapping {
		fmt.Println(block_device_mapping.ID, i)
		block_device_mapping_v2 = append(
			block_device_mapping_v2,
			BlockDeviceMappingV2{
				UUID:            block_device_mapping.ID,
				BootIndex:       block_device_mapping.BootIndex,
				VolumeSize:      block_device_mapping.VolumeSize,
				SourceType:      block_device_mapping.SourceType,
				DestinationType: block_device_mapping.DestinationType,
			})
	}

	payload := RequestPayload{
		Server: Server{
			Name:                 machine.Name,
			KeyName:              machine.Spec.SSHKeyName,
			FlavorRef:            machine.Spec.Flavor,
			AvailabilityZone:     machine.Spec.AvailabilityZone,
			Networks:             networks,
			BlockDeviceMappingV2: block_device_mapping_v2,
		},
	}

	// Marshal the payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return err
	}

	// Define the API URL
	apiURL := Config.ApiBaseURL + Config.Zone + ""

	// Set up the HTTP client
	client := &http.Client{Timeout: 10 * time.Second}

	// Create a new HTTP POST request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return err
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Auth-Token", token) // Replace with your actual token

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return err
	}
	defer resp.Body.Close()

	// Handle the response
	fmt.Println("Response Status:", resp.Status)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return err
	}
	fmt.Println("Response Body:", string(body))

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Println("POST request successful!")
	} else {
		fmt.Println("POST request failed with status:", resp.Status)
	}

	return nil
}
