package plugin

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

func RunPlugin(ctx context.Context, configFlags *genericclioptions.ConfigFlags, outputCh chan string) error {
	config, err := configFlags.ToRESTConfig()
	if err != nil {
		return errors.Wrap(err, "failed to read kubeconfig")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return errors.Wrap(err, "failed to create clientset")
	}

	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to list namespaces")
	}

	for _, namespace := range namespaces.Items {
		configmaps, err := clientset.CoreV1().ConfigMaps(namespace.Name).List(ctx, metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list namespaces")
		}

		for _, configmap := range configmaps.Items {
			outputCh <- fmt.Sprintf("Namespace %s, Configmap %s", namespace.Name, configmap.Name)
		}

	}

	return nil
}
