package cli

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

var cfgFile string

var (
	KubernetesConfigFlags *genericclioptions.ConfigFlags
)

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".directory" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".directory")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

type Options struct {
	genericclioptions.IOStreams

	restConfig *rest.Config

	clientset *kubernetes.Clientset

	namespace string

	Directory string

	Data map[string][]byte

	MetaData MetaData
}

type MetaData struct {
	Annotations map[string]string `yaml:"annotations"`
	Labels      map[string]string `yaml:"labels"`
}

func (o *Options) Setup(f cmdutil.Factory) error {

	config, err := f.ToRESTConfig()
	if err != nil {
		return err
	}

	o.restConfig = config

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	o.clientset = clientset

	namespace, _, err := f.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}

	o.namespace = namespace

	return nil
}

func (o *Options) SetDirectory(pathComponents ...string) {

	o.Directory = filepath.Join(pathComponents...)

}

func (o *Options) EnsureDirectory() error {

	if o.Directory == "" {
		return errors.New("Can't ensure an empty directory")
	}

	return os.MkdirAll(o.Directory, 0777)
}

func (o *Options) SetData(data interface{}) error {

	switch d := data.(type) {
	case map[string][]byte:
		o.Data = d
	case map[string]string:
		convertedData := make(map[string][]byte)
		for key, value := range d {
			convertedData[key] = []byte(value)
		}
		o.Data = convertedData
	default:
		return fmt.Errorf("Can't handle object of type %T", data)
	}

	// Delete the last-applied configuration, because the apiserver manages that for you.
	// Also, it can get pretty big.
	delete(o.MetaData.Annotations, "kubectl.kubernetes.io/last-applied-configuration")

	return nil
}
func (o *Options) WriteData() error {
	err := o.EnsureDirectory()
	if err != nil {
		return err
	}

	metadataFile, err := yaml.Marshal(o.MetaData)
	if err != nil {
		return err
	}

	if len(metadataFile) > 0 {
		err = ioutil.WriteFile(filepath.Join(o.Directory, ".METADATA"), metadataFile, 0644)
		if err != nil {
			return err
		}
	}

	for key, value := range o.Data {
		fmt.Fprintf(o.IOStreams.Out, "Creating %s...", key)

		filename := filepath.Join(o.Directory, key)

		err := ioutil.WriteFile(filename, []byte(value), 0644)
		if err != nil {
			return err
		}
		fmt.Fprint(o.IOStreams.Out, "Done\n")
	}

	return nil
}
