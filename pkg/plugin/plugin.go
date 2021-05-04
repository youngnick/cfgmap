package plugin

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func RunPlugin(ctx context.Context, clientset kubernetes.Clientset, outputCh chan string) error {
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
