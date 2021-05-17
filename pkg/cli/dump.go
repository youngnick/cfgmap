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
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	dumpLong = templates.LongDesc(`
		A kubectl plugin to to dump the contents of a either a ConfigMap
		or Secret to a directory as separate files.
		
		Labels and annotations will be saved in a .metadata.yaml file in the same directory.
				
		The directory will be created as <basedir>/<kind>/<namespace>/<name>, with the
		keys as the filenames. This can be overridden with the --outputdir flag.`)

	dumpExample = templates.Examples(`
		# Export the test configmap from the current namespace
		# to ./configmaps/namespace/test
		%[1]s dump configmap test
		
		# Export the test secret from the current namespace
		# to ./secrets/namespace/test
		%[1]s dump secret test

		# Export the test secret from the test namespace
		# to the default path
		%[1]s dump secret test -n test

		# Export the test secret to ./testsecret
		%[1]s dump secret test --outputdir testsecret
		`)

	parent = "kubectl"
)

func NewDumpCmd(ctx context.Context, in io.Reader, out io.Writer, err io.Writer) *cobra.Command {

	cobra.OnInitialize(initConfig)

	KubernetesConfigFlags = genericclioptions.NewConfigFlags(false)

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	factory := cmdutil.NewFactory(KubernetesConfigFlags)
	ioStreams := genericclioptions.IOStreams{In: in, Out: out, ErrOut: err}
	o := Options{
		IOStreams: ioStreams,
	}

	// dumpCmd represents the dump command
	var dumpCmd = &cobra.Command{
		Use:     fmt.Sprintf("%s dump secret|configmap name", parent),
		Short:   "Dump the contents of an object to a directory as separate files.",
		Long:    fmt.Sprintf(dumpLong, parent),
		Example: fmt.Sprintf(dumpExample, parent),
		Args:    o.ValidateArgumentsRoot,
		Run: func(cmd *cobra.Command, args []string) {
			err := o.Setup(factory)
			cobra.CheckErr(err)

			switch o.Kind {
			case "configmap":
				configmap, err := o.clientset.CoreV1().ConfigMaps(o.Namespace).Get(ctx, o.Name, metav1.GetOptions{})
				cobra.CheckErr(err)
				cobra.CheckErr(o.SetData(configmap.Data))
				o.MetaData.Annotations = configmap.GetObjectMeta().GetAnnotations()
				o.MetaData.Labels = configmap.GetObjectMeta().GetLabels()
			case "secret":
				secret, err := o.clientset.CoreV1().Secrets(o.Namespace).Get(ctx, o.Name, metav1.GetOptions{})
				cobra.CheckErr(err)
				cobra.CheckErr(o.SetData(secret.Data))
				o.MetaData.Annotations = secret.GetObjectMeta().GetAnnotations()
				o.MetaData.Labels = secret.GetObjectMeta().GetLabels()
			}

			// Set up the base directory layout
			configuredOutputDir, err := cmd.Flags().GetString("outputdir")
			cobra.CheckErr(err)
			if configuredOutputDir != "" {
				o.SetDirectory(configuredOutputDir)
			} else {
				basedir, err := cmd.Flags().GetString("basedir")
				cobra.CheckErr(err)
				o.SetDirectory(basedir, o.Kind, o.Namespace, o.Name)

			}

			fmt.Fprintf(ioStreams.Out, "Using directory %s\n", o.Directory)

			cobra.CheckErr(o.WriteData())

		},
	}

	KubernetesConfigFlags.AddFlags(dumpCmd.PersistentFlags())
	dumpCmd.PersistentFlags().String("basedir", ".", "Set the base directory for the configmap directory to be created in.")
	dumpCmd.PersistentFlags().String("outputdir", "", "If supplied, overrides the default directory structure and puts all generated files in the specified directory.")
	return dumpCmd
}
