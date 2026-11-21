package mk8sAddons

// validate.go is the go-libs home of what used to be
// data-service/pkg/handlers/schema_mk8sAddons.go — the mk8s cluster add-on
// config/status sub-schemas embedded by mk8s.ts's spec.addOns and status.addOns
// objects. Each add-on config/status type carries its own ValidationRule (a
// hand-built Object tree, since these exceed the tag DSL) so mk8s.Mk8SClusterSpecAddOns
// / Mk8SStatusAddOns resolve each field through the Validator interface. The byok
// add-on carries the largest sub-tree (a full component-resource config) plus its
// public defaultByokConfig default.
//
// Cpu/Memory come from the workload package (mk8sAddons.ts embeds only Cpu/Memory
// from workload, NOT the workload spec).

import (
	"regexp"

	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/mk8sCommon"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
	"github.com/controlplane-com/libs-go/pkg/schema/workload"
)

// arnRoleRe ports the EFS/ECR/ELB roleArn pattern (mk8sAddons.ts:29 etc.).
var arnRoleRe = regexp.MustCompile(`^arn:(aws):iam::[0-9]+:role/.+`)

func cpuRule() *validate.Rule    { return workload.Cpu("").ValidationRule() }
func memoryRule() *validate.Rule { return workload.Memory("").ValidationRule() }

// ValidationRule ports mk8sAddons.ts:7 RegularExpression (RE2-compile check ==
// base.ts Regex).
func (RegularExpression) ValidationRule() *validate.Rule { return base.RegexRule() }

func regularExpressionRule() *validate.Rule { return base.RegexRule() }

// ValidationRule for the empty add-on config (dashboard/headlamp/awsWorkloadIdentity/
// localPathStorage/sysbox all use NonCustomizableAddonConfig).
func (NonCustomizableAddonConfig) ValidationRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{})
}

func (AzureAddonConfig) ValidationRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"tenantId": validate.String().UUID(),
	})
}

// roleArn is optional since master 405bb677c: when absent the EFS CSI IAM role is created.
func (AwsEFSAddonConfig) ValidationRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"roleArn": validate.String().Regex(arnRoleRe),
	})
}

func (AwsECRAddonConfig) ValidationRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"roleArn": validate.String().Regex(arnRoleRe),
	})
}

func (AwsELBAddonConfig) ValidationRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"roleArn": validate.String().Regex(arnRoleRe),
	})
}

func (AzureACRAddonConfig) ValidationRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"clientId": validate.String().Required(),
	})
}

func (NvidiaAddonConfig) ValidationRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"taintGPUNodes": validate.Boolean(),
	})
}

func (KubevirtAddonConfig) ValidationRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"scratchSpaceStorageClass": validate.String().Optional(),
	})
}

func (NodeLocalDnsAddonConfig) ValidationRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{})
}

func (MetricsAddonConfig) ValidationRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"kubeState":    validate.Boolean(),
		"coreDns":      validate.Boolean(),
		"kubelet":      validate.Boolean(),
		"apiserver":    validate.Boolean(),
		"nodeExporter": validate.Boolean(),
		"cadvisor":     validate.Boolean(),
		"scrapeAnnotated": validate.Object(map[string]*validate.Rule{
			"intervalSeconds":   validate.Number().Integer().Min(30).Default(float64(30)),
			"includeNamespaces": regularExpressionRule(),
			"excludeNamespaces": regularExpressionRule(),
			"retainLabels":      regularExpressionRule(),
		}),
	})
}

func (LogsAddonConfig) ValidationRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"auditEnabled":      validate.Boolean(),
		"includeNamespaces": regularExpressionRule(),
		"excludeNamespaces": regularExpressionRule(),
		"docker":            validate.Boolean(),
		"kubelet":           validate.Boolean(),
		"kernel":            validate.Boolean(),
		"events":            validate.Boolean(),
	})
}

// registryConfigRule ports mk8sAddons.ts:56 RegistryConfig.
func registryConfigRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"registry": validate.Alternatives(mk8sCommon.HostnameStringRule(), validate.String().Valid("*")).Required(),
		"mirrors":  validate.Array(mk8sCommon.UriStringRule()).Max(1),
	})
}

func (RegistryMirrorConfig) ValidationRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"mirrors": validate.Array(registryConfigRule()).Unique(),
	})
}

// --- status add-ons (mk8sAddons.ts:65-110) ---

func (MetricsAddonStatus) ValidationRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"prometheusEndpoint": mk8sCommon.UriStringRule(),
		"remoteWriteConfig":  validate.Object(nil).Unknown(true),
	})
}

func (LogsAddonStatus) ValidationRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"lokiAddress": mk8sCommon.UriStringRule(),
	})
}

func (DashboardAddonStatus) ValidationRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"url": validate.String(),
	})
}

func (AwsWorkloadIdentityAddonStatus) ValidationRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"oidcProviderConfig": validate.Object(map[string]*validate.Rule{
			"providerUrl": mk8sCommon.UriStringRule(),
			"audience":    validate.String(),
		}),
		"trustPolicy": validate.Object(nil).Unknown(true),
	})
}

func (AwsTrustPolicyConfig) ValidationRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"trustPolicy": validate.Object(nil).Unknown(true),
	})
}

// --- prometheus remoteWrite sub-tree (mk8sAddons.ts:146-188) ---

func kvsRule() *validate.Rule {
	return validate.Object(nil).Pattern(validate.String(), validate.String())
}
func tooComplexRule() *validate.Rule { return validate.Object(nil).Unknown(true) }

func httpHeaderValueRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"values":  validate.Array(validate.String()),
		"secrets": validate.Array(validate.String()),
		"files":   validate.Array(validate.String()),
	})
}

// httpConfigKeys returns the base HttpConfig keys (RemoteWrite extends these).
func httpConfigKeys() map[string]*validate.Rule {
	return map[string]*validate.Rule{
		"basic_auth": validate.Object(map[string]*validate.Rule{
			"username":      validate.String(),
			"username_file": validate.String(),
			"password":      validate.String(),
			"password_file": validate.String(),
		}),
		"authorization": validate.Object(map[string]*validate.Rule{
			"type":             validate.String(),
			"credentials":      validate.String(),
			"credentials_file": validate.String(),
		}),
		"oauth2":                 tooComplexRule(),
		"follow_redirects":       validate.Boolean(),
		"enable_http2":           validate.Boolean(),
		"tls_config":             tooComplexRule(),
		"proxy_url":              validate.String(),
		"no_proxy":               validate.String(),
		"proxy_from_environment": validate.Boolean(),
		"proxy_connect_header":   validate.Object(nil).Pattern(validate.String(), validate.Array(validate.String())),
		"http_headers":           validate.Object(nil).Pattern(validate.String(), httpHeaderValueRule()),
	}
}

func remoteWriteRule() *validate.Rule {
	return validate.Object(httpConfigKeys()).Keys(map[string]*validate.Rule{
		"url":                    mk8sCommon.UriStringRule(),
		"remote_timeout":         validate.String(), // Duration
		"headers":                kvsRule(),
		"write_relabel_configs":  validate.Array(tooComplexRule()),
		"name":                   validate.String(),
		"send_exemplars":         validate.Boolean(),
		"send_native_histograms": validate.Boolean(),
		"sigv4":                  tooComplexRule(),
		"azuread":                tooComplexRule(),
		"google_iam":             tooComplexRule(),
		"queue_config":           tooComplexRule(),
	})
}

// --- byok add-on (mk8sAddons.ts:189-306) ---

// byokConfigRule ports the ByokAddonConfig.config nested object.
func byokConfigRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"actuator": validate.Object(map[string]*validate.Rule{
			"minCpu":    cpuRule(),
			"maxCpu":    cpuRule(),
			"minMemory": memoryRule(),
			"maxMemory": memoryRule(),
			"logLevel":  validate.String().Valid("trace", "info", "error"),
			"env":       validate.Object(nil).Pattern(validate.String(), validate.String()),
		}),
		"maintainer": validate.Object(map[string]*validate.Rule{
			"env": validate.Object(nil).Pattern(validate.String(), validate.String()),
		}),
		"juicefs": validate.Object(map[string]*validate.Rule{
			"enabled": validate.Boolean(),
		}),
		"middlebox": validate.Object(map[string]*validate.Rule{
			"enabled":            validate.Boolean(),
			"bandwidthAlertMbps": validate.Number().Integer().Min(0),
			"ip":                 mk8sCommon.IpStringRule([]string{"ipv4"}, "forbidden"),
			"ingressReplicas":    validate.Number().Integer().Min(0).Default(float64(0)),
			"port":               validate.Number(),
		}),
		"common": validate.Object(map[string]*validate.Rule{
			"deploymentReplicas": validate.Number().Integer().Min(0),
			"pdb": validate.Object(map[string]*validate.Rule{
				"maxUnavailable": validate.Number().Integer().Min(0),
			}),
		}),
		"longhorn": validate.Object(map[string]*validate.Rule{
			"numberOfReplicas": validate.Number().Integer().Min(1),
			"replicas":         validate.Number().Optional().Integer().Min(1),
			"isDefault":        validate.Boolean(),
		}),
		"byok": validate.Object(map[string]*validate.Rule{
			"noDefaultStorageClasses": validate.Boolean().Default(false),
		}),
		"ingress": validate.Object(map[string]*validate.Rule{
			"cpu":           cpuRule(),
			"memory":        memoryRule(),
			"targetPercent": validate.Number().Min(0),
		}),
		"istio": validate.Object(map[string]*validate.Rule{
			"istiod": validate.Object(map[string]*validate.Rule{
				"replicas":  validate.Number().Integer().Min(0),
				"minCpu":    cpuRule(),
				"maxCpu":    cpuRule(),
				"minMemory": memoryRule(),
				"maxMemory": memoryRule(),
				"pdb":       validate.Number().Integer().Min(0),
			}),
			"ingressgateway": validate.Object(map[string]*validate.Rule{
				"replicas":  validate.Number().Integer().Min(0),
				"maxCpu":    cpuRule(),
				"maxMemory": memoryRule(),
			}),
			"sidecar": validate.Object(map[string]*validate.Rule{
				"minCpu":    cpuRule(),
				"minMemory": memoryRule(),
			}),
		}),
		"logSplitter": validate.Object(map[string]*validate.Rule{
			"minCpu":        cpuRule(),
			"maxCpu":        cpuRule(),
			"minMemory":     memoryRule(),
			"maxMemory":     memoryRule(),
			"memBufferSize": validate.String(),
			"perPodRate":    validate.Number().Integer().Min(0),
		}),
		"monitoring": validate.Object(map[string]*validate.Rule{
			"minMemory": memoryRule(),
			"maxMemory": memoryRule(),
			"kubeStateMetrics": validate.Object(map[string]*validate.Rule{
				"minMemory": memoryRule(),
			}),
			"prometheus": validate.Object(map[string]*validate.Rule{
				"main": validate.Object(map[string]*validate.Rule{
					"storage": memoryRule(),
				}),
			}),
			"remoteWrite":    validate.Array(remoteWriteRule()),
			"externalLabels": validate.Object(nil).Pattern(validate.String(), validate.String()),
		}),
		"redis": validate.Object(map[string]*validate.Rule{
			"minCpu":    cpuRule(),
			"maxCpu":    cpuRule(),
			"minMemory": memoryRule(),
			"maxMemory": memoryRule(),
			"storage":   memoryRule(),
		}),
		"redisHa": validate.Object(map[string]*validate.Rule{
			"minCpu":    cpuRule(),
			"maxCpu":    cpuRule(),
			"minMemory": memoryRule(),
			"maxMemory": memoryRule(),
			"storage":   validate.Number().Min(0).Integer(),
		}),
		"redisSentinel": validate.Object(map[string]*validate.Rule{
			"minCpu":    cpuRule(),
			"maxCpu":    cpuRule(),
			"minMemory": memoryRule(),
			"maxMemory": memoryRule(),
			"storage":   validate.Number().Min(0).Integer(),
		}),
		"tempoAgent": validate.Object(map[string]*validate.Rule{
			"minCpu":    cpuRule(),
			"minMemory": memoryRule(),
		}),
		"internalDns": validate.Object(map[string]*validate.Rule{
			"minCpu":    cpuRule(),
			"maxCpu":    cpuRule(),
			"minMemory": memoryRule(),
			"maxMemory": memoryRule(),
		}),
	}).Default(func() any { return defaultByokConfigGo() })
}

// ValidationRule ports ByokAddonConfig (mk8sAddons.ts:189). location is a LocalLink
// to a 'location' resource.
func (ByokAddonConfig) ValidationRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"ignoreUpdates": validate.Boolean(),
		"location":      base.LocalLinkRule().Custom(base.ToKind("location")).Required(),
		"config":        byokConfigRule(),
	})
}

// defaultByokConfigGo ports mk8sAddons.ts:309 defaultByokConfig (numbers as
// float64 to match JSON-decoded input and the engine's normalized output).
func defaultByokConfigGo() map[string]any {
	return map[string]any{
		"actuator": map[string]any{
			"minCpu": "50m", "maxCpu": "8001m", "minMemory": "200Mi", "maxMemory": "8000Mi",
			"logLevel": "info",
			"env":      map[string]any{"CACHE_PERIOD_DATA_SERVICE": "600", "LABEL_NODES": "false"},
		},
		"middlebox": map[string]any{"enabled": false, "bandwidthAlertMbps": float64(650)},
		"juicefs":   map[string]any{"enabled": false},
		"common":    map[string]any{"deploymentReplicas": float64(1), "pdb": map[string]any{"maxUnavailable": float64(1)}},
		"longhorn":  map[string]any{"numberOfReplicas": float64(2)},
		"byok":      map[string]any{"noDefaultStorageClasses": false},
		"ingress":   map[string]any{"cpu": "50m", "memory": "200Mi", "targetPercent": float64(6000)},
		"istio": map[string]any{
			"istiod":         map[string]any{"replicas": float64(2), "minCpu": "50m", "maxCpu": "8001m", "minMemory": "100Mi", "maxMemory": "8000Mi", "pdb": float64(0)},
			"ingressgateway": map[string]any{"replicas": float64(2), "maxCpu": "1", "maxMemory": "1Gi"},
			"sidecar":        map[string]any{"minCpu": "0m", "minMemory": "200Mi"},
		},
		"logSplitter": map[string]any{"minCpu": "1m", "maxCpu": "1000m", "minMemory": "10Mi", "maxMemory": "2000Mi", "memBufferSize": "128M", "perPodRate": float64(10000)},
		"monitoring": map[string]any{
			"minMemory": "100Mi", "maxMemory": "20Gi",
			"kubeStateMetrics": map[string]any{"minMemory": "25Mi"},
			"prometheus":       map[string]any{"main": map[string]any{"storage": "10Gi"}},
			"remoteWrite":      []any{},
			"externalLabels":   map[string]any{},
		},
		"redis":         map[string]any{"minCpu": "10m", "maxCpu": "2001m", "minMemory": "100Mi", "maxMemory": "1000Mi", "storage": "8Gi"},
		"redisHa":       map[string]any{"minCpu": "50m", "maxCpu": "2001m", "minMemory": "100Mi", "maxMemory": "1000Mi", "storage": float64(0)},
		"redisSentinel": map[string]any{"minCpu": "10m", "maxCpu": "500m", "minMemory": "10Mi", "maxMemory": "400Mi", "storage": float64(0)},
		"tempoAgent":    map[string]any{"minCpu": "0m", "minMemory": "10Mi"},
	}
}
