package httpapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"dcnlab.ssu.ac.kr/kt-cloud-operator/internal/cloudapi"
	"github.com/kelseyhightower/envconfig"
)

type LoginResponse struct {
	SubjectToken string `json:"subjectToken,omitempty"`
	Token        Token  `json:"token,omitempty"`
	Date         string `json:"date,omitempty"`
}

type Token struct {
	ExpiresAt string `json:"expiresAt,omitempty"`
	IsDomain  bool   `json:"isDomain,omitempty"`
}

// Structs for login
type AuthRequest struct {
	Auth Auth `json:"auth"`
}

type Auth struct {
	Identity Identity `json:"identity"`
	Scope    Scope    `json:"scope"`
}

type Identity struct {
	Methods  []string `json:"methods"`
	Password Password `json:"password"`
}

type Password struct {
	User User `json:"user"`
}

type User struct {
	Domain   Domain `json:"domain"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type Domain struct {
	ID string `json:"id"`
}

type Scope struct {
	Project Project `json:"project"`
}

type Project struct {
	Domain Domain `json:"domain"`
	Name   string `json:"name"`
}

// end structs for login

var Config cloudapi.Config
var logger1 *zap.SugaredLogger

func ProcessEnvVariables() {
	err := envconfig.Process("", &Config)
	if err != nil {
		panic(err.Error())
	}
	err, logger1 = logger(Config.LogLevel)
	if err != nil {
		panic(err.Error())
	}

	logger1.Info("Processed Env Variables...")
}

func KTCloudLogin() {
	ProcessEnvVariables()

	// Create an instance of the struct with your data
	authRequest := AuthRequest{
		Auth: Auth{
			Identity: Identity{
				Methods: []string{Config.IdentityMethods},
				Password: Password{
					User: User{
						Domain:   Domain{ID: Config.IdentityPasswordUserDomainId},
						Name:     Config.IdentityPasswordUserName,
						Password: Config.IdentityPassword,
					},
				},
			},
			Scope: Scope{
				Project: Project{
					Domain: Domain{ID: Config.ScopeProjectDomainId},
					Name:   Config.ScopeProjectName,
				},
			},
		},
	}

	// Marshal the struct to JSON
	payload, err := json.Marshal(authRequest)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}

	// Define the endpoint URL
	apiURL := Config.ApiBaseURL + Config.Zone + "/identity/auth/tokens" // Replace with your API URL

	// Set up HTTP client with timeout
	client := &http.Client{Timeout: 10 * time.Second}

	// Create a new HTTP POST request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payload))
	if err != nil {
		logger1.Fatal("Error creating KT Cloud Auth API request:", err)
		return
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		logger1.Fatal("Error sending KT Cloud Auth POST request:", err)
		return
	}
	defer resp.Body.Close()

	// Handle the response
	logger1.Info("Response Status:", resp.Status)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		logger1.Info("POST request successful!")
		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logger1.Fatal("Error reading response body:", err)
			return
		}

		fmt.Println("Response Headers:")
		for key, values := range resp.Header {
			for _, value := range values {
				// fmt.Printf("%s: %s\n", key, value)
				if key == "X-Subject-Token" {
					logger1.Info("TOKEN: ", value)
				}
			}
		}

		// Print the actual response body
		logger1.Info("Response Body:")
		logger1.Info(string(body))
	} else {
		logger1.Fatal("POST request to KT Cloud Auth failed with status: %s\n", resp.Status)

		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logger1.Fatal("Error reading response body:", err)
			return
		}

		// Print the actual response body
		fmt.Println("Response Body:")
		fmt.Println(string(body))
	}
}

func logger(logLevel string) (error, *zap.SugaredLogger) {
	var level zapcore.Level
	err := level.UnmarshalText([]byte(logLevel))
	if err != nil {
		return err, nil
	}
	logConfig := zap.NewDevelopmentConfig()
	logConfig.Level.SetLevel(level)
	log, err := logConfig.Build()
	if err != nil {
		return err, nil
	}
	return nil, log.Sugar()
}
