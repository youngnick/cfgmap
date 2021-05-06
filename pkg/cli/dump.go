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
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

func NewDumpCmd(ctx context.Context, in io.Reader, out io.Writer, err io.Writer) *cobra.Command {

	// dumpCmd represents the dump command
	var dumpCmd = &cobra.Command{
		Use:   "dump",
		Short: "Dump the contents of objects to a directory as separate files.",
		Long: `This command lets you dump the contents of ConfigMaps or Secrets
		to a directory as separate files.`,
		// Run: func(cmd *cobra.Command, args []string) {}

	}

	cobra.OnInitialize(initConfig)

	KubernetesConfigFlags = genericclioptions.NewConfigFlags(false)
	KubernetesConfigFlags.AddFlags(dumpCmd.PersistentFlags())

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	factory := cmdutil.NewFactory(KubernetesConfigFlags)
	ioStreams := genericclioptions.IOStreams{In: in, Out: out, ErrOut: err}

	dumpCmd.PersistentFlags().String("basedir", ".", "Set the base directory for the configmap directory to be created in.")
	dumpCmd.PersistentFlags().String("outputdir", "", "If supplied, overrides the default directory structure and puts all generated files in the specified directory.")
	dumpCmd.AddCommand(newDumpConfigMapCommand(ctx, ioStreams, factory))
	dumpCmd.AddCommand(newDumpSecretCommand(ctx, ioStreams, factory))
	return dumpCmd
}

func newDumpConfigMapCommand(ctx context.Context, ioStreams genericclioptions.IOStreams, f cmdutil.Factory) *cobra.Command {

	o := Options{
		IOStreams: ioStreams,
	}

	var configMapCmd = &cobra.Command{
		Use:   "configmap",
		Short: "Dump the contents of a ConfigMap to a directory as separate files.",
		Long: `This application is a tool to dump the contents of a ConfigMap
		to a directory as separate files.
		
		The directory will be created as <basedir>/configmaps/<namespace>/<name>, with the
		keys as the filenames.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := o.Setup(f)
			cobra.CheckErr(err)

			name := args[0]
			configmap, err := o.clientset.CoreV1().ConfigMaps(o.namespace).Get(ctx, name, metav1.GetOptions{})
			cobra.CheckErr(err)

			// Set up the base directory layout
			configuredOutputDir, err := cmd.Flags().GetString("outputdir")
			cobra.CheckErr(err)
			if configuredOutputDir != "" {
				o.SetDirectory(configuredOutputDir)
			} else {
				basedir, err := cmd.Flags().GetString("basedir")
				cobra.CheckErr(err)
				o.SetDirectory(basedir, "configmaps", o.namespace, name)

			}

			cobra.CheckErr(o.SetData(configmap.Data))

			fmt.Fprintf(ioStreams.Out, "Using directory %s\n", o.Directory)

			cobra.CheckErr(o.WriteData())

		},
	}

	return configMapCmd
}

func newDumpSecretCommand(ctx context.Context, ioStreams genericclioptions.IOStreams, f cmdutil.Factory) *cobra.Command {

	o := Options{
		IOStreams: ioStreams,
	}

	var secretCmd = &cobra.Command{
		Use:   "secret",
		Short: "Dump the contents of a Secret to a directory as separate files, decoding them on the way.",
		Long: `This application is a tool to dump the contents of a Secret
		to a directory as separate files, decoding them on the way.
		
		By default, the directory will be created as <basedir>/secrets/<namespace>/<name>, with the
		keys as the filenames. This can be overridden with --outputdir.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := o.Setup(f)
			cobra.CheckErr(err)

			name := args[0]
			secret, err := o.clientset.CoreV1().Secrets(o.namespace).Get(ctx, name, metav1.GetOptions{})
			cobra.CheckErr(err)

			// Set up the base directory layout
			configuredOutputDir, err := cmd.Flags().GetString("outputdir")
			cobra.CheckErr(err)
			if configuredOutputDir != "" {
				o.SetDirectory(configuredOutputDir)
			} else {
				basedir, err := cmd.Flags().GetString("basedir")
				cobra.CheckErr(err)
				o.SetDirectory(basedir, "secrets", o.namespace, name)
			}

			cobra.CheckErr(o.SetData(secret.Data))

			fmt.Fprintf(ioStreams.Out, "Using directory %s\n", o.Directory)

			cobra.CheckErr(o.WriteData())

		},
	}

	return secretCmd
}
