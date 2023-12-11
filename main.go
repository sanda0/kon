package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
)

// Server struct to represent server details
type Server struct {
	Name     string `json:"name"`
	IP       string `json:"ip"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func main() {
	// Parse command line arguments
	addNewServer := flag.Bool("n", false, "Add a new server")
	serverName := flag.String("c", "", "Specify server name")
	flag.Parse()

	if *addNewServer {
		addServerInteractive()
		os.Exit(1)
	}

	if *serverName == "" {
		fmt.Println("Please provide a server name using -c flag")
		os.Exit(1)
	}

	// Get the current user
	currentUser, err := user.Current()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Construct the path to the servers.json file using the user's home directory
	filePath := currentUser.HomeDir + "/servers.json"
	// Read server details from JSON file
	servers, err := readServerConfig(filePath)
	if err != nil {
		fmt.Println("Error reading server configuration:", err)
		os.Exit(1)
	}

	// Find the server by name

	server, found := findServerByName(*serverName, servers)
	if !found {
		fmt.Println("Server not found in the configuration")
		os.Exit(1)
	}

	// Build the SSH command
	sshCommand := fmt.Sprintf("sshpass -p '%s' ssh %s@%s", server.Password, server.Username, server.IP)
	fmt.Println("Connecting...")
	// Execute the SSH command
	cmd := exec.Command("bash", "-c", sshCommand)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err = cmd.Run()
	if err != nil {
		fmt.Println("Error executing SSH command:", err)
		os.Exit(1)
	}
}

// Read server configuration from JSON file
func readServerConfig(filename string) ([]Server, error) {
	var servers []Server

	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(fileContent, &servers)
	if err != nil {
		return nil, err
	}

	return servers, nil
}

// Find server by name in the list
func findServerByName(name string, servers []Server) (Server, bool) {
	for _, server := range servers {
		if server.Name == name {
			return server, true
		}
	}
	return Server{}, false
}

// Add a new server interactively and append to the JSON file
func addServerInteractive() {
	var newServer Server

	fmt.Print("Enter server name: ")
	fmt.Scan(&newServer.Name)

	fmt.Print("Enter server IP: ")
	fmt.Scan(&newServer.IP)

	fmt.Print("Enter server username: ")
	fmt.Scan(&newServer.Username)

	fmt.Print("Enter server password: ")
	fmt.Scan(&newServer.Password)

	// Get the current user
	currentUser, err := user.Current()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Construct the path to the servers.json file using the user's home directory
	filePath := currentUser.HomeDir + "/servers.json"

	// Read existing server details from JSON file
	servers, err := readServerConfig(filePath)
	if err != nil {
		// If file doesn't exist, create a new slice of servers
		servers = make([]Server, 0)
	}

	// Append the new server to the existing servers
	servers = append(servers, newServer)

	// Convert servers to JSON
	serverJSON, err := json.MarshalIndent(servers, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling server data:", err)
		os.Exit(1)
	}

	// Write the updated JSON to the file
	err = ioutil.WriteFile(filePath, serverJSON, 0644)
	if err != nil {
		fmt.Println("Error writing to server configuration file:", err)
		os.Exit(1)
	}

	fmt.Println("Server added successfully!")
}
