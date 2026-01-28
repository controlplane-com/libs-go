package leader_election

import (
	"context"
	"sort"

	"github.com/controlplane-com/libs-go/pkg/logging"
	"go.uber.org/zap"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

const (
	TypeLeader   Type = "leader"
	TypeFollower Type = "follower"
	TypeUnknown  Type = "unknown"
)

type Type = string

type Elector interface {
	Role() Type
}

type kubernetesLeaderElector struct {
	api            corev1.CoreV1Interface
	ctx            context.Context
	config         KubernetesLeaderElectionConfig
	kubelessLeader bool
	logger         *zap.SugaredLogger
}

type KubernetesLeaderElectionConfig struct {
	Name        string
	ServiceHost string
	PodName     string
	ServiceName string
	Namespace   string
}

func NewKubernetesLeaderElector(config KubernetesLeaderElectionConfig) Elector {
	ctx := context.Background()
	logger := logging.LoggerWithContext(ctx).With("Name", config.Name, "ServiceHost", config.ServiceHost, "PodName", config.PodName, "ServiceName", config.ServiceName, "Namespace", config.Namespace)
	logger.Infof("starting leader elector")
	// no kube environment, assume leadership without election
	if len(config.ServiceHost) == 0 {
		return &kubernetesLeaderElector{
			ctx:            context.Background(),
			kubelessLeader: true,
			config:         config,
			logger:         logger,
		}
	}
	c, err := rest.InClusterConfig()
	if err != nil {
		logger.Fatalf("Failed to create k8s client: %s", err.Error())
	}
	api, err := kubernetes.NewForConfig(kubernetesTlsConfig(c, config.ServiceHost))
	if err != nil {
		logger.Fatalf("Failed to create k8s client: %s", err.Error())
	}
	e := &kubernetesLeaderElector{
		api:    api.CoreV1(),
		ctx:    context.Background(),
		config: config,
		logger: logger,
	}
	electors[config.Name] = e
	return e
}

func (e *kubernetesLeaderElector) Role() Type {
	if e.kubelessLeader {
		return TypeLeader
	}
	endpoints, err := e.api.Endpoints(e.config.Namespace).
		Get(e.ctx, e.config.ServiceName, metav1.GetOptions{})
	if err != nil {
		e.logger.Warnf("Failed to get endpoints: %s", err.Error())
		return TypeUnknown
	}
	var versions []string
	podMap := map[string]string{}
	for _, subset := range endpoints.Subsets {
		for _, address := range subset.Addresses {
			podMap[address.TargetRef.ResourceVersion] = address.TargetRef.Name
			versions = append(versions, address.TargetRef.ResourceVersion)
		}
	}
	if len(versions) == 0 {
		// should never happen
		e.logger.Warn("No pod name discovered")
		return TypeUnknown
	}
	sort.Strings(versions)
	if podMap[versions[0]] == e.config.PodName {
		return TypeLeader
	}
	return TypeFollower
}

func kubernetesTlsConfig(cfg *rest.Config, kubernetesServiceHost string) *rest.Config {
	if kubernetesServiceHost != "localhost" {
		return cfg
	}
	// allow secure unverified api server transport from local/dev
	cfg.TLSClientConfig.Insecure = true
	cfg.CAFile = ""
	return cfg
}

var electors = map[string]Elector{}

type fixedElector struct {
	role Type
}

func (f *fixedElector) Role() Type {
	return f.role
}

func OverrideElector(class string, role Type) {
	electors[class] = &fixedElector{role: role}
}

var _config KubernetesLeaderElectionConfig

func SetConfig(config KubernetesLeaderElectionConfig) {
	_config = config
}

func LeaderElectorFromConfig(class string) Elector {
	elector, ok := electors[class]
	if !ok {
		elector = NewKubernetesLeaderElector(KubernetesLeaderElectionConfig{
			Name:        class,
			ServiceHost: _config.ServiceHost,
			PodName:     _config.PodName,
			ServiceName: _config.ServiceName,
			Namespace:   _config.Namespace,
		})
		electors[class] = elector
	}
	return elector
}
