package apiserver

import (
	"github.com/etcd-manager/etcd-discovery/apis/etcddiscovery"
	"github.com/etcd-manager/etcd-discovery/apis/etcddiscovery/install"
	"github.com/etcd-manager/etcd-discovery/apis/etcddiscovery/v1alpha1"
	jcstorage "github.com/etcd-manager/etcd-discovery/pkg/registry/etcddiscovery/joincluster"
	pingstorage "github.com/etcd-manager/etcd-discovery/pkg/registry/etcddiscovery/ping"
	"k8s.io/apimachinery/pkg/apimachinery/announced"
	"k8s.io/apimachinery/pkg/apimachinery/registered"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
)

var (
	groupFactoryRegistry = make(announced.APIGroupFactoryRegistry)
	registry             = registered.NewOrDie("")
	Scheme               = runtime.NewScheme()
	Codecs               = serializer.NewCodecFactory(Scheme)
)

func init() {
	install.Install(groupFactoryRegistry, registry, Scheme)

	// we need to add the options to empty v1
	// TODO fix the server code to avoid this
	metav1.AddToGroupVersion(Scheme, schema.GroupVersion{Version: "v1"})

	// TODO: keep the generic API server from wanting this
	unversioned := schema.GroupVersion{Group: "", Version: "v1"}
	Scheme.AddUnversionedTypes(unversioned,
		&metav1.Status{},
		&metav1.APIVersions{},
		&metav1.APIGroupList{},
		&metav1.APIGroup{},
		&metav1.APIResourceList{},
	)
}

type ExtraConfig struct {
	// Place you custom config here.
}

type Config struct {
	GenericConfig *genericapiserver.RecommendedConfig
	ExtraConfig   ExtraConfig
}

// DiscoveryServer contains state for a Kubernetes cluster master/api server.
type DiscoveryServer struct {
	GenericAPIServer *genericapiserver.GenericAPIServer
}

type completedConfig struct {
	GenericConfig genericapiserver.CompletedConfig
	ExtraConfig   *ExtraConfig
}

type CompletedConfig struct {
	// Embed a private pointer that cannot be instantiated outside of this package.
	*completedConfig
}

// Complete fills in any fields not set that are required to have valid data. It's mutating the receiver.
func (cfg *Config) Complete() CompletedConfig {
	c := completedConfig{
		cfg.GenericConfig.Complete(),
		&cfg.ExtraConfig,
	}

	c.GenericConfig.Version = &version.Info{
		Major: "1",
		Minor: "0",
	}

	return CompletedConfig{&c}
}

// New returns a new instance of DiscoveryServer from the given config.
func (c completedConfig) New() (*DiscoveryServer, error) {
	genericServer, err := c.GenericConfig.New("etcd-discovery", genericapiserver.EmptyDelegate)
	if err != nil {
		return nil, err
	}

	s := &DiscoveryServer{
		GenericAPIServer: genericServer,
	}

	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(etcddiscovery.GroupName, registry, Scheme, metav1.ParameterCodec, Codecs)
	apiGroupInfo.GroupMeta.GroupVersion = v1alpha1.SchemeGroupVersion
	v1alpha1storage := map[string]rest.Storage{}
	v1alpha1storage[v1alpha1.ResourcePluralPing] = pingstorage.NewREST()
	v1alpha1storage[v1alpha1.ResourcePluralJoinCluster] = jcstorage.NewREST()
	apiGroupInfo.VersionedResourcesStorageMap[v1alpha1.SchemeGroupVersion.Version] = v1alpha1storage

	if err := s.GenericAPIServer.InstallAPIGroup(&apiGroupInfo); err != nil {
		return nil, err
	}

	return s, nil
}
