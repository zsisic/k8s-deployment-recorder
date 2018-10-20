package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/sirupsen/logrus"
)

func main() {
	logrus.Info("hello world")

	config := getConfiguration()

	logrus.Println("Configuration: ", config)

	clientset := getClietnSet(config.kubeconfig)

	namespaces := getNamespaces(clientset, config.namespaces, config.excludeNamespaces)

	logrus.Println("Monitored namespaces: ", namespaces)

	records := getDeploymentRecordsForNamespaces(clientset, namespaces)

	logrus.Println("deployments: ", records)
}

func getDeploymentRecordsForNamespaces(clientset *kubernetes.Clientset, namespaces []string) (deploymentRecords []DeploymentRecord) {

	deploymentClient := clientset.AppsV1()

	for _, ns := range namespaces {
		deployments, err := deploymentClient.Deployments(ns).List(metav1.ListOptions{})
		verifyNoError(err)
		for _, dItem := range deployments.Items {
			deploymentRecords = append(deploymentRecords, *buildDeploymentRecord(&dItem))
		}
	}
	return
}

func buildDeploymentRecord(item *v1.Deployment) *DeploymentRecord {

	var dockerImages []string
	for _, c := range item.Spec.Template.Spec.Containers {
		dockerImages = append(dockerImages, c.Image)
	}

	deploymentRecor := DeploymentRecord{
		cluster:           item.GetClusterName(),
		namespace:         item.GetNamespace(),
		applicationName:   item.GetLabels()["app"],
		deloymentName:     item.GetName(),
		deploymentVersion: item.GetLabels()["version"],
		dockerImages:      dockerImages,
	}

	return &deploymentRecor
}

func getNamespaces(clientset *kubernetes.Clientset, requestedNamespaces, namespacesToExclude []string) (namespaces []string) {

	namespaceInclusionTable := map[string]bool{}
	for _, ns := range requestedNamespaces {
		namespaceInclusionTable[ns] = true
	}
	for _, ns := range namespacesToExclude {
		namespaceInclusionTable[ns] = false
	}

	namespaceItems, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})
	verifyNoError(err)
	logrus.Println("retrieved items,", namespaceItems.Items)
	logrus.Println("Configuration is: ", namespaceInclusionTable)
	for _, item := range namespaceItems.Items {
		normalizedNs := strings.ToLower(item.Name)
		if len(namespaceInclusionTable) > 0 {
			shouldInclude, found := namespaceInclusionTable[normalizedNs]
			logrus.Println("ns evaluation ", normalizedNs, found, shouldInclude)
			if found && !shouldInclude {
				continue
			}
		}
		namespaces = append(namespaces, item.Name)
	}

	return
}
func getConfiguration() *Configuration {
	var kubeconfig, namespaces, excludedNamespaces *string
	var refreshFrequency *int

	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	namespaces = flag.String("namespaces", "", "comma-delimited list of namespaces to monitor, all if not specified")
	excludedNamespaces = flag.String("excludedNamespaces", "default,kube-public,kube-system", "comma-delimited list of namespace to exclude from monitoring")
	refreshFrequency = flag.Int("refreshFrequency", 30, "Poll/Refresh frequency in seconds")

	flag.Parse()

	return &Configuration{
		kubeconfig:        *kubeconfig,
		namespaces:        splitList(*namespaces),
		excludeNamespaces: splitList(*excludedNamespaces),
		refreshFrequency:  *refreshFrequency,
	}
}

func splitList(str string) []string {
	var trimmed = strings.TrimSpace(str)
	trimmed = strings.ToLower(trimmed)
	return strings.Split(trimmed, ",")
}

func getClietnSet(kubeconfig string) *kubernetes.Clientset {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	verifyNoError(err)

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	verifyNoError(err)
	return clientset
}

func verifyNoError(err error) {
	if err != nil {
		panic(err.Error())
	}
}
func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
