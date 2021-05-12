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
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/cli-runtime/pkg/genericclioptions"
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

	loadCmd.PersistentFlags().String("basedir", ".", "Set the base directory for the default search path, <basedir>/<object>/<namespace>/<name>.")
	loadCmd.PersistentFlags().String("inputdir", "", "If supplied, overrides the default search path and loads all files from the specified directory.")
	loadCmd.AddCommand(newLoadConfigMapCommand(ctx, ioStreams, factory))
	loadCmd.AddCommand(newLoadSecretCommand(ctx, ioStreams, factory))
	return loadCmd
}

func newLoadConfigMapCommand(ctx context.Context, ioStreams genericclioptions.IOStreams, f cmdutil.Factory) *cobra.Command {

	o := Options{
		IOStreams: ioStreams,
	}

	var configMapCmd = &cobra.Command{
		Use:   "configmap <name>",
		Short: "Load the contents of a ConfigMap from a directory as separate files.",
		Long: `This application is a tool to load the contents of a ConfigMap
		from a directory containing files, the names of which will be used as the keys.

		Labels and annotations may be read from a .metadat.yaml file in the source directory.
		
		The default source directory is <basedir>/configmaps/<namespace>/<name>, with the
		keys as the filenames.
		
		This behavior can be overridden with the --inputDir flag.`,
		Args: o.ValidateArguments,
		Run: func(cmd *cobra.Command, args []string) {
			err := o.Setup(f)
			cobra.CheckErr(err)

			inputDir, err := cmd.Flags().GetString("inputdir")
			cobra.CheckErr(err)
			if inputDir != "" {
				o.SetDirectory(inputDir)
			} else {
				basedir, err := cmd.Flags().GetString("basedir")
				cobra.CheckErr(err)

				o.SetDirectory(basedir, "configmaps", o.Namespace, o.Name)
			}
			err = o.ReadData()
			cobra.CheckErr(err)
			configmap := o.GetConfigMap()

			o.Printer, err = genericclioptions.NewJSONYamlPrintFlags().ToPrinter("yaml")
			cobra.CheckErr(err)

			err = o.Printer.PrintObj(configmap, o.Out)
			cobra.CheckErr(err)

		},
	}

	return configMapCmd
}

func newLoadSecretCommand(ctx context.Context, ioStreams genericclioptions.IOStreams, f cmdutil.Factory) *cobra.Command {

	o := Options{
		IOStreams: ioStreams,
	}

	var secretCmd = &cobra.Command{
		Use:   "secret <name>",
		Short: "Load the contents of a Secret to a directory as separate files, decoding them on the way.",
		Long: `This application is a tool to dump the contents of a Secret
		from a directory containing files, the names of which will be used as the keys.

		Labels and annotations may be read from a .metadata.yaml file in the source directory.
		
		The default source directory is <basedir>/secrets/<namespace>/<name>, with the
		keys as the filenames.
		
		This behavior can be overridden with the --inputdir flag.`,
		Args: o.ValidateArguments,
		Run: func(cmd *cobra.Command, args []string) {
			err := o.Setup(f)
			cobra.CheckErr(err)

			inputDir, err := cmd.Flags().GetString("inputdir")
			cobra.CheckErr(err)
			if inputDir != "" {
				o.SetDirectory(inputDir)
			} else {
				basedir, err := cmd.Flags().GetString("basedir")
				cobra.CheckErr(err)

				o.SetDirectory(basedir, "secrets", o.Namespace, o.Name)
			}

			err = o.ReadData()
			cobra.CheckErr(err)
			secret := o.GetSecret()

			o.Printer, err = genericclioptions.NewJSONYamlPrintFlags().ToPrinter("yaml")
			cobra.CheckErr(err)

			err = o.Printer.PrintObj(secret, o.Out)
			cobra.CheckErr(err)
		},
	}

	return secretCmd
}
