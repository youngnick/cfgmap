/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cli

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

func newDumpCmd(ctx context.Context, ioStreams genericclioptions.IOStreams, f cmdutil.Factory) *cobra.Command {

	// dumpCmd represents the dump command
	var dumpCmd = &cobra.Command{
		Use:   "dump",
		Short: "Dump the contents of a ConfigMap to a directory as separate files.",
		Long: `This application is a tool to dump the contents of a ConfigMap
		to a directory as separate files.
		
		The directory will be created as <basedir>/<namespace>/<name>, with the
		keys as the filenames.`,
		Run: func(cmd *cobra.Command, args []string) {
			config, err := f.ToRESTConfig()
			cobra.CheckErr(err)

			clientset, err := kubernetes.NewForConfig(config)
			cobra.CheckErr(err)

			namespace, _, err := f.ToRawKubeConfigLoader().Namespace()
			cobra.CheckErr(err)
			name := args[0]
			configmap, err := clientset.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
			cobra.CheckErr(err)
			// fmt.Fprint(ioStreams.Out, configmap)

			basedir, err := cmd.Flags().GetString("basedir")
			cobra.CheckErr(err)

			fmt.Fprintf(ioStreams.Out, "Using directory %s/%s/%s\n", basedir, namespace, name)
			errMkdir := os.MkdirAll(fmt.Sprintf("%s/%s/%s", basedir, namespace, name), 0777)
			cobra.CheckErr(errMkdir)

			for key, value := range configmap.Data {
				fmt.Fprintf(ioStreams.Out, "Creating %s...", key)

				filename := filepath.Join(basedir, namespace, name, key)

				err := ioutil.WriteFile(filename, []byte(value), 0644)
				cobra.CheckErr(err)
				fmt.Fprint(ioStreams.Out, "Done\n")
			}
			// for _, namespace := range namespaces.Items {
			// 	configmaps, err := clientset.CoreV1().ConfigMaps(namespace.Name).List(ctx, metav1.ListOptions{})
			// 	if err != nil {
			// 		cobra.CheckErr(err)
			// 	}

			// 	for _, configmap := range configmaps.Items {

			// 	}

			// }

			// clientset, err :=
		},
	}

	dumpCmd.Flags().String("basedir", ".", "Set the base directory for the configmap directory to be created in.")
	return dumpCmd
}
