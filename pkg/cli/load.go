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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	loadLong = templates.LongDesc(`
		A kubectl plugin to to load the contents of a either a ConfigMap
		or Secret from a directory of separate files.
		
		Labels and annotations will be loaded from a .metadata.yaml file in the same
		directory, if it exists.
				
		The default search path is (basedir)/(kind)/(namespace)/(name), and the filenames
		will be used for the keys in the ConfigMap or Secret.
		
		The default search path can be overridden with the --inputdir flag.`)

	loadExample = templates.Examples(`
		# Load the "test" configmap into the "namespace" namespace
		# from directory ./configmaps/namespace/test
		%[1]s load configmap test
		
		# Load the "test" Secret into the "namespace" namespace
		# from directory ./configmaps/namespace/test
		%[1]s load secret test

		# Load the "test" Secret into the "test" namespace
		# from directory ./configmaps/test/test
		%[1]s load secret test -n test

		# Load the test secret into the current namespace
		# from ./testsecret
		%[1]s dump secret test --inputputdir testsecret
		`)
)

func NewLoadCmd(ctx context.Context, in io.Reader, out io.Writer, err io.Writer) *cobra.Command {
	cobra.OnInitialize(initConfig)

	KubernetesConfigFlags = genericclioptions.NewConfigFlags(false)

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	factory := cmdutil.NewFactory(KubernetesConfigFlags)
	ioStreams := genericclioptions.IOStreams{In: in, Out: out, ErrOut: err}

	o := Options{
		IOStreams: ioStreams,
	}

	// loadCmd represents the load command
	var loadCmd = &cobra.Command{
		Use:     fmt.Sprintf("%s load secret|configmap name", parentCmd),
		Short:   "Load the contents of an object from the files in a directory, with the filenames as keys.",
		Long:    fmt.Sprintf(loadLong),
		Example: fmt.Sprintf(loadExample, parentCmd),
		Args:    o.ValidateArgumentsRoot,

		Run: func(cmd *cobra.Command, args []string) {
			err := o.Setup(factory)
			cobra.CheckErr(err)

			var typeDir string

			switch o.Kind {
			case "configmap":
				typeDir = "configmaps"
			case "secret":
				typeDir = "secrets"
			}

			inputDir, err := cmd.Flags().GetString("inputdir")
			cobra.CheckErr(err)
			if inputDir != "" {
				o.SetDirectory(inputDir)
			} else {
				basedir, err := cmd.Flags().GetString("basedir")
				cobra.CheckErr(err)

				o.SetDirectory(basedir, typeDir, o.Namespace, o.Name)
			}
			err = o.ReadData()
			cobra.CheckErr(err)

			var obj runtime.Object

			switch o.Kind {
			case "configmap":
				obj = o.GetConfigMap()
			case "secret":
				obj = o.GetSecret()
			}

			o.Printer, err = genericclioptions.NewJSONYamlPrintFlags().ToPrinter("yaml")
			cobra.CheckErr(err)

			err = o.Printer.PrintObj(obj, o.Out)
			cobra.CheckErr(err)

		},
	}

	KubernetesConfigFlags.AddFlags(loadCmd.PersistentFlags())

	loadCmd.PersistentFlags().String("basedir", ".", "Set the base directory for the default search path, <basedir>/<object>/<namespace>/<name>.")
	loadCmd.PersistentFlags().String("inputdir", "", "If supplied, overrides the default search path and loads all files from the specified directory.")

	// Hide most of the standard Kubectl flags. They'll still work, just not show up in help.
	loadCmd.PersistentFlags().MarkHidden("as-group")
	loadCmd.PersistentFlags().MarkHidden("as")
	loadCmd.PersistentFlags().MarkHidden("cache-dir")
	loadCmd.PersistentFlags().MarkHidden("certificate-authority")
	loadCmd.PersistentFlags().MarkHidden("client-certificate")
	loadCmd.PersistentFlags().MarkHidden("client-key")
	loadCmd.PersistentFlags().MarkHidden("cluster")
	loadCmd.PersistentFlags().MarkHidden("context")
	loadCmd.PersistentFlags().MarkHidden("insecure-skip-tls-verify")
	loadCmd.PersistentFlags().MarkHidden("kubeconfig")
	loadCmd.PersistentFlags().MarkHidden("password")
	loadCmd.PersistentFlags().MarkHidden("request-timeout")
	loadCmd.PersistentFlags().MarkHidden("server")
	loadCmd.PersistentFlags().MarkHidden("tls-server-name")
	loadCmd.PersistentFlags().MarkHidden("token")
	loadCmd.PersistentFlags().MarkHidden("user")
	loadCmd.PersistentFlags().MarkHidden("username")
	return loadCmd
}
