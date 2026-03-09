package provider

import (
	"context"
	"fmt"
	"log"

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
	client    kubernetes.Interface
	namespace string
}

func NewKubernetesProvider(kubeconfig, namespace string) (*KubernetesProvider, error) {
	var config *rest.Config
	var err error

	// Try in-cluster config first, fall back to kubeconfig
	config, err = rest.InClusterConfig()
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
		client:    clientset,
		namespace: namespace,
	}, nil
}

func (k *KubernetesProvider) CreateInstance(input *CreateInstanceInput) error {
	ctx := context.Background()
	inst := input.Instance
	tmpl := input.Template
	resourceName := inst.DeploymentName // claw-{id}

	// 1. Create Secret
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceName,
			Namespace: k.namespace,
			Labels:    k.labels(inst.ID),
		},
		StringData: map[string]string{
			"api_key":  input.APIKey,
			"mm_token": input.MMBotToken,
		},
	}
	if _, err := k.client.CoreV1().Secrets(k.namespace).Create(ctx, secret, metav1.CreateOptions{}); err != nil {
		return fmt.Errorf("create secret: %w", err)
	}
	log.Printf("[K8s] Created Secret %s/%s", k.namespace, resourceName)

	// 2. Create Deployment
	replicas := int32(1)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceName,
			Namespace: k.namespace,
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
	if _, err := k.client.AppsV1().Deployments(k.namespace).Create(ctx, deployment, metav1.CreateOptions{}); err != nil {
		return fmt.Errorf("create deployment: %w", err)
	}
	log.Printf("[K8s] Created Deployment %s/%s", k.namespace, resourceName)

	// 3. Create Service
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceName,
			Namespace: k.namespace,
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
	if _, err := k.client.CoreV1().Services(k.namespace).Create(ctx, service, metav1.CreateOptions{}); err != nil {
		return fmt.Errorf("create service: %w", err)
	}
	log.Printf("[K8s] Created Service %s/%s", k.namespace, resourceName)

	return nil
}

func (k *KubernetesProvider) DeleteInstance(instance *models.Instance) error {
	ctx := context.Background()
	resourceName := instance.DeploymentName

	// Delete in reverse order: Service, Deployment, Secret
	if err := k.client.CoreV1().Services(k.namespace).Delete(ctx, resourceName, metav1.DeleteOptions{}); err != nil {
		log.Printf("[K8s] Warning: delete service %s: %v", resourceName, err)
	}
	if err := k.client.AppsV1().Deployments(k.namespace).Delete(ctx, resourceName, metav1.DeleteOptions{}); err != nil {
		log.Printf("[K8s] Warning: delete deployment %s: %v", resourceName, err)
	}
	if err := k.client.CoreV1().Secrets(k.namespace).Delete(ctx, resourceName, metav1.DeleteOptions{}); err != nil {
		log.Printf("[K8s] Warning: delete secret %s: %v", resourceName, err)
	}

	log.Printf("[K8s] Deleted resources for instance %s", instance.ID)
	return nil
}

func (k *KubernetesProvider) GetInstanceStatus(instance *models.Instance) (*InstanceStatus, error) {
	ctx := context.Background()
	resourceName := instance.DeploymentName

	deployment, err := k.client.AppsV1().Deployments(k.namespace).Get(ctx, resourceName, metav1.GetOptions{})
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
	pods, err := k.client.CoreV1().Pods(k.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("lobsterpool.io/instance=%s", instance.ID),
	})
	if err == nil && len(pods.Items) > 0 {
		for _, cs := range pods.Items[0].Status.ContainerStatuses {
			if cs.State.Waiting != nil && cs.State.Waiting.Reason == "CrashLoopBackOff" {
				status = "failed"
			}
		}
	}

	endpoint := fmt.Sprintf("http://%s.%s.svc.cluster.local", resourceName, k.namespace)

	return &InstanceStatus{
		Status:   status,
		Endpoint: endpoint,
	}, nil
}

func (k *KubernetesProvider) labels(instanceID string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/managed-by": "lobsterpool",
		"lobsterpool.io/instance":      instanceID,
	}
}
