package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"dcnlab.ssu.ac.kr/kt-cloud-operator/api/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Meta API for object metadata

// type KTMachineStatus struct {
// 	ID             string           `json:"id"`
// 	AdminPass      string           `json:"adminPass"`
// 	Links          []Links          `json:"links"`
// 	SecurityGroups []SecurityGroups `json:"securityGroups"`
// }

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

type NetworkTier struct {
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
	Networks             []NetworkTier          `json:"networks"`
	BlockDeviceMappingV2 []BlockDeviceMappingV2 `json:"block_device_mapping_v2"`
}

type RequestPayload struct {
	Server Server `json:"server"`
}

// Define the struct to parse the response
type ServerResponse struct {
	Server v1beta1.KTMachineStatus `json:"server"`
}

func CreateVM(machine *v1beta1.KTMachine, token string) error {
	// Create the payload
	networks := []NetworkTier{}
	block_device_mapping_v2 := []BlockDeviceMappingV2{}

	for i, network := range machine.Spec.NetworkTier {
		fmt.Println(network.ID, i)
		networks = append(
			networks,
			NetworkTier{
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
		logger1.Error("Error marshaling JSON for POST machine creation:", err)
		return err
	}

	// Define the API URL
	apiURL := Config.ApiBaseURL + Config.Zone + "/server/servers"

	// Set up the HTTP client
	client := &http.Client{Timeout: 10 * time.Second}

	// Create a new HTTP POST request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
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
		logger1.Info("POST request successful and created machine!")

		// Parse the JSON into the struct
		var serverResponse ServerResponse
		err = json.Unmarshal(body, &serverResponse)
		if err != nil {
			logger1.Error("Error unmarshaling JSON response:", err)
			return err
		}

		// Access the parsed data
		// fmt.Println("Parsed Response:")
		// fmt.Printf("Server ID: %s\n", serverResponse.Server.ID)
		// fmt.Printf("Admin Password: %s\n", serverResponse.Server.AdminPass)
		// fmt.Printf("Disk Config: %s\n", serverResponse.Server.DiskConfig)
		// fmt.Println("Links:")
		for _, link := range serverResponse.Server.Links {
			fmt.Printf("  - Rel: %s, Href: %s\n", link.Rel, link.Href)
		}
		fmt.Println("Security Groups:")
		for _, group := range serverResponse.Server.SecurityGroups {
			fmt.Printf("  - Name: %s\n", group.Name)
		}

		// Update the machine K8s Resource
		clientConfig, err := getRestConfig(Config.Kubeconfig)
		if err != nil {
			logger1.Errorf("Cannot prepare k8s client config: %v. Kubeconfig was: %s", err, Config.Kubeconfig)
			panic(err.Error())
		}
		// Set up a scheme (use runtime.Scheme from apimachinery)
		scheme := runtime.NewScheme()
		// Create Kubernetes client
		k8sClient, err := getClient(clientConfig, scheme)
		if err != nil {
			logger1.Fatalf("Failed to create Kubernetes client: %v", err)
			return err
		}

		// machineStatus := &v1beta1.KTMachineStatus{
		// 	ID: serverResponse.Server.ID,
		// 	AdminPass: serverResponse.Server.AdminPass,
		// 	Links: serverResponse.Server.Links,
		// }
		err = updateVMStatus(k8sClient, machine, &serverResponse.Server)
		if err != nil {
			logger1.Errorf("Failed to update VMstatus: %v", err)
			return err
		}
		logger1.Info("Updated the status of machine")
		return nil

	} else {
		logger1.Error("POST request failed with status:", resp.Status)
	}

	return nil
}
func updateVMStatus(k8sClient client.Client, machine *v1beta1.KTMachine, newMachineStatus *v1beta1.KTMachineStatus) error {
	ctx := context.Background()

	machine.Status = *newMachineStatus
	machine.Status.State = "Creating"
	err := k8sClient.Status().Update(ctx, machine)
	if err != nil {
		logger1.Errorf("Failed to update KTSubjectToken object: %v", err)
		return err
	}
	return nil
}
