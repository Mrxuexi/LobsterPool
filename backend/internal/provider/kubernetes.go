package provider

import (
	"context"
	"fmt"
	"log"
	"sort"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/lobsterpool/lobsterpool/internal/models"
)

type KubernetesProvider struct {
	clusters       map[string]*clusterClient
	clusterOrder   []string
	defaultCluster string
}

type clusterClient struct {
	info   ClusterInfo
	client kubernetes.Interface
}

func NewLegacyKubernetesProvider(kubeconfig, namespace, defaultCluster string) (*KubernetesProvider, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("build kubeconfig: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("create kubernetes client: %w", err)
	}

	return &KubernetesProvider{
		clusters: map[string]*clusterClient{
			defaultCluster: {
				info: ClusterInfo{
					Name:        defaultCluster,
					DisplayName: defaultCluster,
					Namespace:   namespace,
					Default:     true,
				},
				client: clientset,
			},
		},
		clusterOrder:   []string{defaultCluster},
		defaultCluster: defaultCluster,
	}, nil
}

func NewKubernetesProvider(clusters []ClusterConfig, defaultCluster string) (*KubernetesProvider, error) {
	if len(clusters) == 0 {
		return nil, fmt.Errorf("no kubernetes clusters configured")
	}

	provider := &KubernetesProvider{
		clusters:       make(map[string]*clusterClient, len(clusters)),
		clusterOrder:   make([]string, 0, len(clusters)),
		defaultCluster: defaultCluster,
	}

	for _, cluster := range clusters {
		restConfig, err := buildClusterRESTConfig(cluster)
		if err != nil {
			return nil, fmt.Errorf("build cluster %q config: %w", cluster.Name, err)
		}

		clientset, err := kubernetes.NewForConfig(restConfig)
		if err != nil {
			return nil, fmt.Errorf("create cluster %q client: %w", cluster.Name, err)
		}

		provider.clusters[cluster.Name] = &clusterClient{
			info: ClusterInfo{
				Name:        cluster.Name,
				DisplayName: cluster.DisplayName,
				Namespace:   cluster.Namespace,
				Default:     cluster.Name == defaultCluster,
			},
			client: clientset,
		}
		provider.clusterOrder = append(provider.clusterOrder, cluster.Name)
	}

	sort.Strings(provider.clusterOrder)
	return provider, nil
}

func buildClusterRESTConfig(cluster ClusterConfig) (*rest.Config, error) {
	if cluster.APIServer != "" {
		return &rest.Config{
			Host:        cluster.APIServer,
			BearerToken: cluster.Token,
			TLSClientConfig: rest.TLSClientConfig{
				CAFile:   cluster.CAFile,
				Insecure: cluster.InsecureSkipTLSVerify,
			},
		}, nil
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if cluster.Kubeconfig != "" {
		loadingRules.ExplicitPath = cluster.Kubeconfig
	}

	overrides := &clientcmd.ConfigOverrides{}
	if cluster.Context != "" {
		overrides.CurrentContext = cluster.Context
	}

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("load kubeconfig: %w", err)
	}
	return config, nil
}

func (k *KubernetesProvider) clusterForInstance(instance *models.Instance) (*clusterClient, error) {
	clusterName := instance.Cluster
	if clusterName == "" {
		clusterName = k.defaultCluster
	}

	cluster, ok := k.clusters[clusterName]
	if !ok {
		return nil, fmt.Errorf("cluster %q is not configured", clusterName)
	}
	return cluster, nil
}

func (k *KubernetesProvider) CreateInstance(input *CreateInstanceInput) error {
	ctx := context.Background()
	inst := input.Instance
	tmpl := input.Template
	resourceName := inst.DeploymentName // claw-{id}
	cluster, err := k.clusterForInstance(inst)
	if err != nil {
		return err
	}
	namespace := inst.Namespace
	if namespace == "" {
		namespace = cluster.info.Namespace
	}

	// 1. Create Secret
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceName,
			Namespace: namespace,
			Labels:    k.labels(inst.ID),
		},
		StringData: map[string]string{
			"api_key":  input.APIKey,
			"mm_token": input.MMBotToken,
		},
	}
	if _, err := cluster.client.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{}); err != nil {
		return fmt.Errorf("create secret: %w", err)
	}
	log.Printf("[K8s] Created Secret %s/%s on cluster %s", namespace, resourceName, cluster.info.Name)

	// 2. Create Deployment
	replicas := int32(1)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceName,
			Namespace: namespace,
			Labels:    k.labels(inst.ID),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: k.labels(inst.ID),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: k.labels(inst.ID),
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "openclaw",
							Image: fmt.Sprintf("%s:%s", tmpl.Image, tmpl.Version),
							Ports: []corev1.ContainerPort{
								{ContainerPort: int32(tmpl.DefaultPort)},
							},
							Env: []corev1.EnvVar{
								{
									Name: "OPENAI_API_KEY",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{Name: resourceName},
											Key:                  "api_key",
										},
									},
								},
								{
									Name: "MATTERMOST_BOT_TOKEN",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{Name: resourceName},
											Key:                  "mm_token",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	if _, err := cluster.client.AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{}); err != nil {
		return fmt.Errorf("create deployment: %w", err)
	}
	log.Printf("[K8s] Created Deployment %s/%s on cluster %s", namespace, resourceName, cluster.info.Name)

	// 3. Create Service
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceName,
			Namespace: namespace,
			Labels:    k.labels(inst.ID),
		},
		Spec: corev1.ServiceSpec{
			Selector: k.labels(inst.ID),
			Ports: []corev1.ServicePort{
				{
					Port:       int32(tmpl.DefaultPort),
					TargetPort: intstr.FromInt(tmpl.DefaultPort),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
	if _, err := cluster.client.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{}); err != nil {
		return fmt.Errorf("create service: %w", err)
	}
	log.Printf("[K8s] Created Service %s/%s on cluster %s", namespace, resourceName, cluster.info.Name)

	return nil
}

func (k *KubernetesProvider) DeleteInstance(instance *models.Instance) error {
	ctx := context.Background()
	resourceName := instance.DeploymentName
	cluster, err := k.clusterForInstance(instance)
	if err != nil {
		return err
	}
	namespace := instance.Namespace
	if namespace == "" {
		namespace = cluster.info.Namespace
	}

	// Delete in reverse order: Service, Deployment, Secret
	if err := cluster.client.CoreV1().Services(namespace).Delete(ctx, resourceName, metav1.DeleteOptions{}); err != nil {
		log.Printf("[K8s] Warning: delete service %s: %v", resourceName, err)
	}
	if err := cluster.client.AppsV1().Deployments(namespace).Delete(ctx, resourceName, metav1.DeleteOptions{}); err != nil {
		log.Printf("[K8s] Warning: delete deployment %s: %v", resourceName, err)
	}
	if err := cluster.client.CoreV1().Secrets(namespace).Delete(ctx, resourceName, metav1.DeleteOptions{}); err != nil {
		log.Printf("[K8s] Warning: delete secret %s: %v", resourceName, err)
	}

	log.Printf("[K8s] Deleted resources for instance %s on cluster %s", instance.ID, cluster.info.Name)
	return nil
}

func (k *KubernetesProvider) GetInstanceStatus(instance *models.Instance) (*InstanceStatus, error) {
	ctx := context.Background()
	resourceName := instance.DeploymentName
	cluster, err := k.clusterForInstance(instance)
	if err != nil {
		return nil, err
	}
	namespace := instance.Namespace
	if namespace == "" {
		namespace = cluster.info.Namespace
	}

	deployment, err := cluster.client.AppsV1().Deployments(namespace).Get(ctx, resourceName, metav1.GetOptions{})
	if err != nil {
		return &InstanceStatus{Status: "not_found"}, nil
	}

	status := "pending"
	if deployment.Status.ReadyReplicas > 0 {
		status = "running"
	} else if deployment.Status.UnavailableReplicas > 0 {
		status = "starting"
	}

	// Check for crash/error
	pods, err := cluster.client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("lobsterpool.io/instance=%s", instance.ID),
	})
	if err == nil && len(pods.Items) > 0 {
		for _, cs := range pods.Items[0].Status.ContainerStatuses {
			if cs.State.Waiting != nil && cs.State.Waiting.Reason == "CrashLoopBackOff" {
				status = "failed"
			}
		}
	}

	endpoint := fmt.Sprintf("http://%s.%s.svc.cluster.local", resourceName, namespace)

	return &InstanceStatus{
		Status:   status,
		Endpoint: endpoint,
	}, nil
}

func (k *KubernetesProvider) ListClusters() []ClusterInfo {
	clusters := make([]ClusterInfo, 0, len(k.clusterOrder))
	for _, name := range k.clusterOrder {
		clusters = append(clusters, k.clusters[name].info)
	}
	return clusters
}

func (k *KubernetesProvider) labels(instanceID string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/managed-by": "lobsterpool",
		"lobsterpool.io/instance":      instanceID,
	}
}
