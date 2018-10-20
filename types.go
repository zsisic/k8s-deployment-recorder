package main

// Configuration as application configuration data
type Configuration struct {
	kubeconfig        string
	namespaces        []string
	excludeNamespaces []string
	refreshFrequency  int
}

// DeploymentRecord prvides sumary of a deployment
type DeploymentRecord struct {
	cluster           string
	namespace         string
	deloymentName     string
	deploymentVersion string
	applicationName   string
	dockerImages      []string
}
