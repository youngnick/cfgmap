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
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

func NewLoadCmd(ctx context.Context, in io.Reader, out io.Writer, err io.Writer) *cobra.Command {

	// loadCmd represents the load command
	var loadCmd = &cobra.Command{
		Use:   "load",
		Short: "Load the contents of objects to a directory as separate files.",
		Long: `This command lets you load the contents of ConfigMaps or Secrets
		to a directory as separate files.`,
		// Run: func(cmd *cobra.Command, args []string) {}

	}

	cobra.OnInitialize(initConfig)

	KubernetesConfigFlags = genericclioptions.NewConfigFlags(false)
	KubernetesConfigFlags.AddFlags(loadCmd.PersistentFlags())

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	factory := cmdutil.NewFactory(KubernetesConfigFlags)
	ioStreams := genericclioptions.IOStreams{In: in, Out: out, ErrOut: err}

	loadCmd.PersistentFlags().String("basedir", ".", "Set the base directory to start looking for the configmap directory.")
	loadCmd.PersistentFlags().String("outputdir", "", "If supplied, overrides the default directory structure and puts all generated files in the specified directory.")
	loadCmd.AddCommand(newLoadConfigMapCommand(ctx, ioStreams, factory))
	loadCmd.AddCommand(newLoadSecretCommand(ctx, ioStreams, factory))
	return loadCmd
}

func newLoadConfigMapCommand(ctx context.Context, ioStreams genericclioptions.IOStreams, f cmdutil.Factory) *cobra.Command {

	var configMapCmd = &cobra.Command{
		Use:   "configmap",
		Short: "Load the contents of a ConfigMap to a directory as separate files.",
		Long: `This application is a tool to dump the contents of a ConfigMap
		to a directory as separate files.
		
		The directory will be created as <basedir>/configmaps/<namespace>/<name>, with the
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

			basedir, err := cmd.Flags().GetString("basedir")
			cobra.CheckErr(err)

			// Set up the base directory layout
			outputDir := filepath.Join(basedir, "configmaps", namespace, name)
			configuredOutputDir, err := cmd.Flags().GetString("outputdir")
			cobra.CheckErr(err)
			if configuredOutputDir != "" {
				outputDir = configuredOutputDir
			}
			fmt.Fprintf(ioStreams.Out, "Using directory %s\n", outputDir)
			errMkdir := os.MkdirAll(outputDir, 0777)
			cobra.CheckErr(errMkdir)

			for key, value := range configmap.Data {
				fmt.Fprintf(ioStreams.Out, "Creating %s...", key)

				filename := filepath.Join(outputDir, key)

				err := ioutil.WriteFile(filename, []byte(value), 0644)
				cobra.CheckErr(err)
				fmt.Fprint(ioStreams.Out, "Done\n")
			}

		},
	}

	return configMapCmd
}

func newLoadSecretCommand(ctx context.Context, ioStreams genericclioptions.IOStreams, f cmdutil.Factory) *cobra.Command {

	var secretCmd = &cobra.Command{
		Use:   "secret",
		Short: "Load the contents of a Secret to a directory as separate files, decoding them on the way.",
		Long: `This application is a tool to dump the contents of a Secret
		to a directory as separate files, decoding them on the way.
		
		By default, the directory will be created as <basedir>/secrets/<namespace>/<name>, with the
		keys as the filenames. This can be overridden with --outputdir.`,
		Run: func(cmd *cobra.Command, args []string) {
			config, err := f.ToRESTConfig()
			cobra.CheckErr(err)

			clientset, err := kubernetes.NewForConfig(config)
			cobra.CheckErr(err)

			namespace, _, err := f.ToRawKubeConfigLoader().Namespace()
			cobra.CheckErr(err)
			name := args[0]
			secret, err := clientset.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
			cobra.CheckErr(err)

			basedir, err := cmd.Flags().GetString("basedir")
			cobra.CheckErr(err)

			// Set up the base directory layout
			outputDir := filepath.Join(basedir, "secrets", namespace, name)
			configuredOutputDir, err := cmd.Flags().GetString("outputdir")
			cobra.CheckErr(err)
			if configuredOutputDir != "" {
				outputDir = configuredOutputDir
			}
			fmt.Fprintf(ioStreams.Out, "Using directory %s\n", outputDir)
			errMkdir := os.MkdirAll(outputDir, 0777)
			cobra.CheckErr(errMkdir)

			for key, value := range secret.Data {
				fmt.Fprintf(ioStreams.Out, "Creating %s...", key)

				filename := filepath.Join(outputDir, key)

				err := ioutil.WriteFile(filename, []byte(value), 0644)
				cobra.CheckErr(err)
				fmt.Fprint(ioStreams.Out, "Done\n")
			}

		},
	}

	return secretCmd
}
