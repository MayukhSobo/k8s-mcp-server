package kubernetes

import (
	"context"
	"fmt"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// Client represents a Kubernetes client
type Client struct {
	clientset     *kubernetes.Clientset
	dynamicClient dynamic.Interface
}

// NewClient creates a new Kubernetes client
func NewClient(kubeconfigPath string) (*Client, error) {
	var config *rest.Config
	var err error

	if kubeconfigPath == "" {
		// Try in-cluster config
		config, err = rest.InClusterConfig()
		if err != nil {
			// If not in cluster, try default kubeconfig path
			home := homedir.HomeDir()
			if home != "" {
				kubeconfigPath = filepath.Join(home, ".kube", "config")
			}
		}
	}

	if config == nil {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to build kubeconfig: %v", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes clientset: %v", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %v", err)
	}

	return &Client{
		clientset:     clientset,
		dynamicClient: dynamicClient,
	}, nil
}

// GetClientset returns the Kubernetes clientset
func (c *Client) GetClientset() *kubernetes.Clientset {
	return c.clientset
}

// GetResource retrieves a specific resource by name
func (c *Client) GetResource(resourceType, namespace, name string) (*unstructured.Unstructured, error) {
	gvr, err := getGroupVersionResource(resourceType)
	if err != nil {
		return nil, err
	}

	var resource *unstructured.Unstructured
	if namespace != "" {
		resource, err = c.dynamicClient.Resource(gvr).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	} else {
		resource, err = c.dynamicClient.Resource(gvr).Get(context.TODO(), name, metav1.GetOptions{})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get %s '%s': %v", resourceType, name, err)
	}

	return resource, nil
}

// ListResources lists resources of a specific type
func (c *Client) ListResources(resourceType, namespace string) (*unstructured.UnstructuredList, error) {
	gvr, err := getGroupVersionResource(resourceType)
	if err != nil {
		return nil, err
	}

	var resources *unstructured.UnstructuredList
	if namespace != "" {
		resources, err = c.dynamicClient.Resource(gvr).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
	} else {
		resources, err = c.dynamicClient.Resource(gvr).List(context.TODO(), metav1.ListOptions{})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list %s: %v", resourceType, err)
	}

	return resources, nil
}

// CreateResource creates a new resource
func (c *Client) CreateResource(resourceType, namespace string, object *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	gvr, err := getGroupVersionResource(resourceType)
	if err != nil {
		return nil, err
	}

	var created *unstructured.Unstructured
	if namespace != "" {
		created, err = c.dynamicClient.Resource(gvr).Namespace(namespace).Create(context.TODO(), object, metav1.CreateOptions{})
	} else {
		created, err = c.dynamicClient.Resource(gvr).Create(context.TODO(), object, metav1.CreateOptions{})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create %s: %v", resourceType, err)
	}

	return created, nil
}

// DeleteResource deletes a resource
func (c *Client) DeleteResource(resourceType, namespace, name string) error {
	gvr, err := getGroupVersionResource(resourceType)
	if err != nil {
		return err
	}

	var deleteErr error
	if namespace != "" {
		deleteErr = c.dynamicClient.Resource(gvr).Namespace(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	} else {
		deleteErr = c.dynamicClient.Resource(gvr).Delete(context.TODO(), name, metav1.DeleteOptions{})
	}

	if deleteErr != nil {
		return fmt.Errorf("failed to delete %s '%s': %v", resourceType, name, deleteErr)
	}

	return nil
}

// getGroupVersionResource maps a resource type to its GroupVersionResource
func getGroupVersionResource(resourceType string) (schema.GroupVersionResource, error) {
	// Map of common resource types to their GroupVersionResource
	resourceMap := map[string]schema.GroupVersionResource{
		"pods":                   {Group: "", Version: "v1", Resource: "pods"},
		"services":               {Group: "", Version: "v1", Resource: "services"},
		"deployments":            {Group: "apps", Version: "v1", Resource: "deployments"},
		"namespaces":             {Group: "", Version: "v1", Resource: "namespaces"},
		"configmaps":             {Group: "", Version: "v1", Resource: "configmaps"},
		"secrets":                {Group: "", Version: "v1", Resource: "secrets"},
		"persistentvolumes":      {Group: "", Version: "v1", Resource: "persistentvolumes"},
		"persistentvolumeclaims": {Group: "", Version: "v1", Resource: "persistentvolumeclaims"},
		"statefulsets":           {Group: "apps", Version: "v1", Resource: "statefulsets"},
		"daemonsets":             {Group: "apps", Version: "v1", Resource: "daemonsets"},
		"ingresses":              {Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"},
	}

	gvr, exists := resourceMap[resourceType]
	if !exists {
		return schema.GroupVersionResource{}, fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	return gvr, nil
}
