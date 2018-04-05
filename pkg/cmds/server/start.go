package server

import (
	"fmt"
	"io"
	"net"

	"github.com/etcd-manager/etcd-discovery/apis/etcddiscovery/v1alpha1"
	"github.com/etcd-manager/etcd-discovery/pkg/apiserver"
	"github.com/spf13/pflag"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	genericapiserver "k8s.io/apiserver/pkg/server"
	genericoptions "k8s.io/apiserver/pkg/server/options"
)

const defaultEtcdPathPrefix = "/registry/etcddiscovery.pharmer.com"

type DiscoveryServerOptions struct {
	RecommendedOptions *genericoptions.RecommendedOptions
	StdOut             io.Writer
	StdErr             io.Writer
}

func NewDiscoveryServerOptions(out, errOut io.Writer) *DiscoveryServerOptions {
	o := &DiscoveryServerOptions{
		RecommendedOptions: genericoptions.NewRecommendedOptions(defaultEtcdPathPrefix, apiserver.Codecs.LegacyCodec(v1alpha1.SchemeGroupVersion)),
		StdOut:             out,
		StdErr:             errOut,
	}
	o.RecommendedOptions.Etcd = nil
	o.RecommendedOptions.SecureServing.BindPort = 8443
	o.RecommendedOptions.Authorization = nil
	return o
}

func (o *DiscoveryServerOptions) AddFlags(fs *pflag.FlagSet) {
	o.RecommendedOptions.AddFlags(fs)
}

func (o DiscoveryServerOptions) Validate(args []string) error {
	var errors []error
	errors = append(errors, o.RecommendedOptions.Validate()...)
	return utilerrors.NewAggregate(errors)
}

func (o *DiscoveryServerOptions) Complete() error {
	return nil
}

func (o DiscoveryServerOptions) Config() (*apiserver.Config, error) {
	// TODO have a "real" external address
	if err := o.RecommendedOptions.SecureServing.MaybeDefaultWithSelfSignedCerts("localhost", nil, []net.IP{net.ParseIP("127.0.0.1")}); err != nil {
		return nil, fmt.Errorf("error creating self-signed certificates: %v", err)
	}

	serverConfig := genericapiserver.NewRecommendedConfig(apiserver.Codecs)
	if err := o.RecommendedOptions.ApplyTo(serverConfig); err != nil {
		return nil, err
	}

	config := &apiserver.Config{
		GenericConfig: serverConfig,
		ExtraConfig:   apiserver.ExtraConfig{},
	}
	return config, nil
}

func (o DiscoveryServerOptions) Run(stopCh <-chan struct{}) error {
	config, err := o.Config()
	if err != nil {
		return err
	}

	server, err := config.Complete().New()
	if err != nil {
		return err
	}

	server.GenericAPIServer.AddPostStartHook("start-etcd-discovery-server-informers", func(context genericapiserver.PostStartHookContext) error {
		config.GenericConfig.SharedInformerFactory.Start(context.StopCh)
		return nil
	})

	return server.GenericAPIServer.PrepareRun().Run(stopCh)
}
