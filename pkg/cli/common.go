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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
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

	Name      string
	Namespace string

	Directory string

	Data map[string][]byte

	MetaData MetaData

	Printer printers.ResourcePrinter
}

type MetaData struct {
	Annotations map[string]string `yaml:"annotations,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
}

// Setup initialises the RestConfig, Clientset, and namespace fields.
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

	if o.Namespace == "" {
		namespace, _, err := f.ToRawKubeConfigLoader().Namespace()
		if err != nil {
			return err
		}
		o.Namespace = namespace
	}

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
		err = ioutil.WriteFile(filepath.Join(o.Directory, ".metadata.yaml"), metadataFile, 0644)
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

func (o *Options) ReadData() error {

	files, err := os.ReadDir(o.Directory)
	if err != nil {
		return err
	}

	o.Data = make(map[string][]byte)

	for _, file := range files {

		filename := file.Name()
		rawData, err := os.ReadFile(filepath.Join(o.Directory, filename))
		if err != nil {
			return err
		}
		if filename == ".metadata.yaml" {
			err = yaml.Unmarshal(rawData, &o.MetaData)
			if err != nil {
				return err
			}
			continue
		}

		o.Data[filename] = rawData

	}
	return nil
}

func (o *Options) GetSecret() *v1.Secret {

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        o.Name,
			Namespace:   o.Namespace,
			Annotations: o.MetaData.Annotations,
			Labels:      o.MetaData.Labels,
		},
		Data: o.Data,
	}

	secret.SetGroupVersionKind(v1.SchemeGroupVersion.WithKind("Secret"))
	return secret
}

func (o *Options) GetConfigMap() *v1.ConfigMap {

	data := make(map[string]string)

	for key, value := range o.Data {
		data[key] = string(value)
	}

	cfgmap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        o.Name,
			Namespace:   o.Namespace,
			Annotations: o.MetaData.Annotations,
			Labels:      o.MetaData.Labels,
		},
		Data: data,
	}
	cfgmap.SetGroupVersionKind(v1.SchemeGroupVersion.WithKind("ConfigMap"))

	return cfgmap
}

func (o *Options) ValidateArguments(cmd *cobra.Command, args []string) error {

	if len(args) > 1 {
		return errors.New("Only name is used")
	}

	if len(args) < 1 {
		return errors.New("Name is required")
	}

	o.Name = args[0]

	return nil
}
