package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	corev1alpha1 "kubesphere.io/api/core/v1alpha1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubesphere/ksbuilder/pkg/utils"
)

type unpublishOptions struct{}

func defaultUnpublishOptions() *unpublishOptions {
	return &unpublishOptions{}
}

func unpublishExtensionCmd() *cobra.Command {
	o := defaultUnpublishOptions()

	cmd := &cobra.Command{
		Use:   "unpublish",
		Short: "Unpublish an extension from the market",
		Args:  cobra.ExactArgs(1),
		RunE:  o.unpublish,
	}
	return cmd
}

func (o *unpublishOptions) unpublish(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	name := args[0]
	fmt.Printf("unpublish extension %s\n", name)

	flagVal, _ := cmd.Root().PersistentFlags().GetString("kubeconfig")
	kubeconfigPath := utils.ResolveKubeconfig(flagVal)
	fmt.Printf("Using kubeconfig: %s\n", kubeconfigPath)
	genericClient, err := utils.BuildClientFromFlags(kubeconfigPath)
	if err != nil {
		return err
	}

	extensionVersions := &corev1alpha1.ExtensionVersionList{}
	if err = genericClient.List(ctx, extensionVersions, runtimeclient.MatchingLabels{
		corev1alpha1.ExtensionReferenceLabel: name,
	}); err != nil {
		return err
	}

	installPlan := &corev1alpha1.InstallPlan{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kubesphere.io/v1alpha1",
			Kind:       "InstallPlan",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	if err := deleteObjAndWait(ctx, genericClient, installPlan, time.Minute, 2*time.Second); err != nil {
		return err
	}

	objs := make([]runtimeclient.Object, 0)
	for i := range extensionVersions.Items {
		version := &extensionVersions.Items[i]
		objs = append(objs, &corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "ConfigMap",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("extension-%s-chart", version.Name),
				Namespace: "kubesphere-system",
			},
		}, version)
	}

	return deleteObjs(ctx, genericClient, append(objs, &corev1alpha1.Extension{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kubesphere.io/v1alpha1",
			Kind:       "Extension",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	})...)
}

func deleteObjs(ctx context.Context, c runtimeclient.Client, objs ...runtimeclient.Object) error {
	for _, obj := range objs {
		fmt.Printf("deleting %s %s\n", obj.GetObjectKind().GroupVersionKind().Kind, obj.GetName())
		if err := c.Delete(ctx, obj); err != nil && !errors.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func deleteObjAndWait(ctx context.Context, c runtimeclient.Client, obj runtimeclient.Object, timeout, interval time.Duration) error {
	fmt.Printf("deleting %s %s\n", obj.GetObjectKind().GroupVersionKind().Kind, obj.GetName())
	if err := c.Delete(ctx, obj); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	key := runtimeclient.ObjectKeyFromObject(obj)
	current, ok := obj.DeepCopyObject().(runtimeclient.Object)
	if !ok {
		return fmt.Errorf("object %T does not implement client.Object", obj)
	}

	return wait.PollUntilContextTimeout(ctx, interval, timeout, true, func(ctx context.Context) (bool, error) {
		if err := c.Get(ctx, key, current); err != nil {
			if errors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		}
		return false, nil
	})
}
