package workspace

import (
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v3"
)

func getWorkspaceFilePath() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(home, ".config", "petitorium")
	return filepath.Join(configDir, "workspace.yaml"), nil
}

func createDefaultWorkspace() *Workspace {
	return &Workspace{
		Collections: []Collection{
			{
				Name: "Users",
				Requests: []Request{
					{
						Name:   "index",
						Method: "GET",
						URL:    "https://dummyjson.com/users",
					},
					{
						Name:   "show",
						Method: "GET",
						URL:    "{{base_url}}/users/1",
					},
					{
						Name:   "store",
						Method: "POST",
						URL:    "https://jsonplaceholder.org/users",
						Body: `{
									"firstName": "James",
									"lastName": "Davis",
									"maidenName": "",
									"age": 45,
									"gender": "male",
									"email": "james.davis@x.dummyjson.com",
									"phone": "+49 614-958-9364",
									"username": "jamesd",
									"password": "jamesdpass",
									"birthDate": "1979-5-4",
									"image": "https://dummyjson.com/icon/jamesd/128",
									"bloodGroup": "AB+",
									"height": 193.31,
									"weight": 62.1,
									"eyeColor": "Amber",
									"hair": {
										"color": "Blonde",
										"type": "Straight"
									},
									"ip": "101.118.131.66",
									"address": {
										"address": "238 Jefferson Street",
										"city": "Seattle",
										"state": "Pennsylvania",
										"stateCode": "PA",
										"postalCode": "68354",
										"coordinates": {
											"lat": 16.782513,
											"lng": -139.34723
										},
										"country": "United States"
									},
									"macAddress": "10:7d:df:1f:97:58",
									"university": "University of Southern California",
									"bank": {
										"cardExpire": "05/29",
										"cardNumber": "5005519846254763",
										"cardType": "Mastercard",
										"currency": "INR",
										"iban": "7N7ZH1PJ8Q4WU1K965HQQR27"
									},
									"company": {
										"department": "Support",
										"name": "Pagac and Sons",
										"title": "Research Analyst",
										"address": {
											"address": "1622 Lincoln Street",
											"city": "Fort Worth",
											"state": "Pennsylvania",
											"stateCode": "PA",
											"postalCode": "27768",
											"coordinates": {
												"lat": 54.91193,
												"lng": -79.498328
											},
											"country": "United States"
										}
									},
									"ein": "904-810",
									"ssn": "116-951-314",
									"userAgent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.99 Safari/537.36",
									"crypto": {
										"coin": "Bitcoin",
										"wallet": "0xb9fc2fe63b2a6c003f1c324c3bfa53259162181a",
										"network": "Ethereum (ERC20)"
									},
									"role": "admin"
							}`,
					},
					{
						Name:   "update",
						Method: "PUT",
						URL:    "https://dummyjson.com/users/2",
						Body: `{
									"lastName": "Owais"
							}`,
					},
					{
						Name:   "delete",
						Method: "DELETE",
						URL:    "https://dummyjson.com/users/1",
					},
				},
				Collections: []Collection{
					{
						Name: "Others",
						Requests: []Request{
							{
								Name:   "update",
								Method: "PATCH",
								URL:    "https://dummyjson.com/users/2",
								Body: `{
									"lastName": "Owais"
							}`,
							},
							{
								Name:   "head",
								Method: "HEAD",
								URL:    "https://dummyjson.com/users",
							},
							{
								Name:   "options",
								Method: "OPTIONS",
								URL:    "https://dummyjson.com/users",
							},
						},
					},
				},
			},
			{
				Name: "Posts",
				Requests: []Request{
					{
						Name:   "show",
						Method: "GET",
						URL:    "'https://dummyjson.com/posts/1'",
					},
				},
			},
		},
		Requests: []Request{
			{
				Name:   "Health Check",
				Method: "GET",
				URL:    "https://httpbin.org/status/200",
			},
			{
				Name:   "Echo",
				Method: "POST",
				URL:    "https://httpbin.org/post",
				Body:   `{"message": "Hello World"}`,
			},
		},
	}
}

func LoadWorkspace() (*Workspace, error) {
	path, err := getWorkspaceFilePath()
	if err != nil {
		return nil, err
	}

	if _, err = os.Stat(path); os.IsNotExist(err) {
		defaultWorkspace := createDefaultWorkspace()
		if err = SaveWorkspace(defaultWorkspace); err != nil {
			return nil, err
		}
		return defaultWorkspace, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var workspace Workspace
	if err := yaml.Unmarshal(data, &workspace); err != nil {
		return nil, err
	}

	// Load expansion state if available
	if err := LoadExpansionState(&workspace.Collections); err != nil {
		// If there's an error loading expansion state, continue without it
		// This is not a critical error
	}

	return &workspace, nil
}

func SaveWorkspace(workspace *Workspace) error {
	path, err := getWorkspaceFilePath()
	if err != nil {
		return err
	}

	if err = os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := yaml.Marshal(workspace)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

// Helper function to get expansion state file path
func getEnvironmentsFilePath() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(home, ".config", "petitorium")
	return filepath.Join(configDir, "environments.yaml"), nil
}

func createDefaultEnvironments() []Environment {
	return []Environment{
		{
			Name: "Base",
			Variables: map[string]string{
				"base_url": "https://api.example.com",
			},
		},
	}
}

// GetEffectiveVariables returns the effective variables for an environment,
// merging with base environment variables if specified
func (e *Environment) GetEffectiveVariables(environments []Environment) map[string]string {
	effective := make(map[string]string)

	// First, inherit from base environment if specified
	if e.Base != "" {
		for _, env := range environments {
			if env.Name == e.Base {
				// Recursively get base variables (to handle multiple levels of inheritance)
				baseVars := env.GetEffectiveVariables(environments)
				for k, v := range baseVars {
					effective[k] = v
				}
				break
			}
		}
	}

	// Then override with this environment's variables
	for k, v := range e.Variables {
		effective[k] = v
	}

	return effective
}

func LoadEnvironments() ([]Environment, error) {
	path, err := getEnvironmentsFilePath()
	if err != nil {
		return nil, err
	}

	if _, err = os.Stat(path); os.IsNotExist(err) {
		defaultEnvs := createDefaultEnvironments()
		if err = SaveEnvironments(defaultEnvs); err != nil {
			return nil, err
		}
		return defaultEnvs, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var environments []Environment
	if err := yaml.Unmarshal(data, &environments); err != nil {
		return nil, err
	}

	return environments, nil
}

func SaveEnvironments(environments []Environment) error {
	path, err := getEnvironmentsFilePath()
	if err != nil {
		return err
	}

	if err = os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := yaml.Marshal(environments)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

func getExpansionStateFilePath() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(home, ".config", "petitorium")
	return filepath.Join(configDir, "expansion_state.yaml"), nil
}

// SaveExpansionState saves the expansion state of collections
func SaveExpansionState(collections *[]Collection) error {
	path, err := getExpansionStateFilePath()
	if err != nil {
		return err
	}

	if err = os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	// Create a simple map for expansion state
	expansionState := make(map[string]bool)

	var collectExpansionState func(collections *[]Collection, prefix string)
	collectExpansionState = func(collections *[]Collection, prefix string) {
		for _, col := range *collections {
			fullName := prefix + col.Name
			expansionState[fullName] = col.Expanded
			if len(col.Collections) > 0 {
				collectExpansionState(&col.Collections, fullName+"/")
			}
		}
	}

	collectExpansionState(collections, "")

	data, err := yaml.Marshal(expansionState)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

// LoadExpansionState loads the expansion state and applies it to collections
func LoadExpansionState(collections *[]Collection) error {
	path, err := getExpansionStateFilePath()
	if err != nil {
		return err
	}

	if _, err = os.Stat(path); os.IsNotExist(err) {
		// No expansion state file exists yet, that's fine
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var expansionState map[string]bool
	if err := yaml.Unmarshal(data, &expansionState); err != nil {
		return err
	}

	var applyExpansionState func(collections *[]Collection, prefix string)
	applyExpansionState = func(collections *[]Collection, prefix string) {
		for i := range *collections {
			fullName := prefix + (*collections)[i].Name
			if expanded, exists := expansionState[fullName]; exists {
				(*collections)[i].Expanded = expanded
			}
			if len((*collections)[i].Collections) > 0 {
				applyExpansionState(&(*collections)[i].Collections, fullName+"/")
			}
		}
	}

	applyExpansionState(collections, "")
	return nil
}
