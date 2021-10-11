package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"sort"
	"strings"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/yaml"
)

const (
	separatorGV    = "/"
	separatorPanic = ": "
	separatorYAML  = "---\n"

	verbGet   = "get"
	verbList  = "list"
	verbWatch = "watch"

	coreGroupName      = ""
	secretResourceName = "secrets"

	namespacedReadOnlyRoleName    = "namespaced-ro"
	namespacedReadWriteRoleName   = "namespaced-rw"
	unnamespacedReadOnlyRoleName  = "unnamespaced-ro"
	unnamespacedReadWriteRoleName = "unnamespaced-rw"
)

var (
	readOnlyVerbs = []string{
		verbGet,
		verbList,
		verbWatch,
	}

	readWriteVerbs = []string{
		rbacv1.VerbAll,
	}
)

type ClusterRoles struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	KubeConfig        ClusterRolesKubeConfig `json:"kubeConfig,omitempty"`
}

type ClusterRolesKubeConfig struct {
	LoadingRules *clientcmd.ClientConfigLoadingRules `json:"loadingRules,omitempty"`
	Overrides    *clientcmd.ConfigOverrides          `json:"overrides,omitempty"`
}

func main() {
	filePath := os.Args[1]

	loadingRules, overrides, err := readClientConfigSettings(filePath)
	if err != nil {
		log.Panic(filePath, separatorPanic, err)
	}

	deferredLoadingClientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides)
	clientConfig, err := deferredLoadingClientConfig.ClientConfig()
	if err != nil {
		log.Panic(filePath, separatorPanic, err)
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(clientConfig)
	if err != nil {
		log.Panic(filePath, separatorPanic, err)
	}

	index, err := buildIndex(discoveryClient)
	if err != nil {
		log.Panic(filePath, separatorPanic, err)
	}

	clusterRoles, err := makeClusterRoles(index)
	if err != nil {
		log.Panic(filePath, separatorPanic, err)
	}
	canonicalizeClusterRoles(clusterRoles)

	for _, clusterRole := range clusterRoles {
		bytes, err := yaml.Marshal(clusterRole)
		if err != nil {
			log.Panic(filePath, separatorPanic, err)
		}

		if _, err := os.Stdout.Write(bytes); err != nil {
			log.Panic(filePath, separatorPanic, err)
		}

		if _, err := os.Stdout.Write([]byte(separatorYAML)); err != nil {
			log.Panic(filePath, separatorPanic, err)
		}
	}
}

func readClientConfigSettings(filePath string) (*clientcmd.ClientConfigLoadingRules, *clientcmd.ConfigOverrides, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, nil, err
	}

	clusterRoles := ClusterRoles{
		KubeConfig: ClusterRolesKubeConfig{
			LoadingRules: clientcmd.NewDefaultClientConfigLoadingRules(),
			Overrides:    &clientcmd.ConfigOverrides{},
		},
	}
	if err := yaml.Unmarshal(data, &clusterRoles); err != nil {
		return nil, nil, err
	}

	return clusterRoles.KubeConfig.LoadingRules, clusterRoles.KubeConfig.Overrides, nil
}

type Namespaced bool
type ResourceIndex map[string]Namespaced
type GroupIndex map[string]ResourceIndex

func buildIndex(discoveryClient *discovery.DiscoveryClient) (GroupIndex, error) {
	_, resourceLists, err := discoveryClient.ServerGroupsAndResources()
	if err != nil {
		return nil, err
	}

	groupIndex := make(GroupIndex)
	for _, resourceList := range resourceLists {
		groupVersion := resourceList.GroupVersion

		var groupName string
		if separatorIndex := strings.Index(groupVersion, separatorGV); separatorIndex != -1 {
			groupName = groupVersion[:separatorIndex]
		}

		resourceIndex, ok := groupIndex[groupName]
		if !ok {
			resourceIndex = make(ResourceIndex)
			groupIndex[groupName] = resourceIndex
		}

		for _, resource := range resourceList.APIResources {
			resourceIndex[resource.Name] = Namespaced(resource.Namespaced)
		}
	}

	return groupIndex, nil
}

func makeClusterRoles(index GroupIndex) ([]rbacv1.ClusterRole, error) {
	var clusterRoles []rbacv1.ClusterRole

	namespacedRoles, err := makeNamespacedClusterRoles(index)
	if err != nil {
		return nil, err
	}
	clusterRoles = append(clusterRoles, namespacedRoles...)

	unnamespacedRoles, err := makeUnnamespacedClusterRoles(index)
	if err != nil {
		return nil, err
	}
	clusterRoles = append(clusterRoles, unnamespacedRoles...)

	return clusterRoles, nil
}

func makeNamespacedClusterRoles(index GroupIndex) ([]rbacv1.ClusterRole, error) {
	coreRule := rbacv1.PolicyRule{
		APIGroups: []string{
			coreGroupName,
		},
		Verbs: readOnlyVerbs,
	}

	othersRule := rbacv1.PolicyRule{
		Resources: []string{
			rbacv1.ResourceAll,
		},
		Verbs: readOnlyVerbs,
	}

	for group, resources := range index {
		if group != coreGroupName {
			othersRule.APIGroups = append(othersRule.APIGroups, group)
			continue
		}

		for resource, namespaced := range resources {
			if resource != secretResourceName && namespaced {
				coreRule.Resources = append(coreRule.Resources, resource)
			}
		}
	}

	typeMeta := metav1.TypeMeta{
		APIVersion: rbacv1.SchemeGroupVersion.String(),
		Kind:       reflect.TypeOf(rbacv1.ClusterRole{}).Name(),
	}
	clusterRoles := []rbacv1.ClusterRole{
		rbacv1.ClusterRole{
			TypeMeta: typeMeta,
			ObjectMeta: metav1.ObjectMeta{
				Name: namespacedReadOnlyRoleName,
			},
			Rules: []rbacv1.PolicyRule{
				coreRule,
				othersRule,
			},
		},
		rbacv1.ClusterRole{
			TypeMeta: typeMeta,
			ObjectMeta: metav1.ObjectMeta{
				Name: namespacedReadWriteRoleName,
			},
			Rules: []rbacv1.PolicyRule{
				rbacv1.PolicyRule{
					APIGroups: []string{
						rbacv1.APIGroupAll,
					},
					Resources: []string{
						rbacv1.ResourceAll,
					},
					Verbs: readWriteVerbs,
				},
			},
		},
	}
	return clusterRoles, nil
}

func makeUnnamespacedClusterRoles(index GroupIndex) ([]rbacv1.ClusterRole, error) {
	var readOnlyRules []rbacv1.PolicyRule
	var readWriteRules []rbacv1.PolicyRule
	for group, resources := range index {
		var unnamespacedResources []string
		for resource, namespaced := range resources {
			if !namespaced {
				unnamespacedResources = append(unnamespacedResources, resource)
			}
		}
		if len(unnamespacedResources) == 0 {
			continue
		}

		groups := []string{
			group,
		}

		readOnlyRules = append(readOnlyRules, rbacv1.PolicyRule{
			APIGroups: groups,
			Resources: unnamespacedResources,
			Verbs:     readOnlyVerbs,
		})

		readWriteRules = append(readWriteRules, rbacv1.PolicyRule{
			APIGroups: groups,
			Resources: unnamespacedResources,
			Verbs:     readWriteVerbs,
		})
	}

	typeMeta := metav1.TypeMeta{
		APIVersion: rbacv1.SchemeGroupVersion.String(),
		Kind:       reflect.TypeOf(rbacv1.ClusterRole{}).Name(),
	}
	clusterRoles := []rbacv1.ClusterRole{
		rbacv1.ClusterRole{
			TypeMeta: typeMeta,
			ObjectMeta: metav1.ObjectMeta{
				Name: unnamespacedReadOnlyRoleName,
			},
			Rules: readOnlyRules,
		},
		rbacv1.ClusterRole{
			TypeMeta: typeMeta,
			ObjectMeta: metav1.ObjectMeta{
				Name: unnamespacedReadWriteRoleName,
			},
			Rules: readWriteRules,
		},
	}
	return clusterRoles, nil
}

func canonicalizeClusterRoles(clusterRoles []rbacv1.ClusterRole) {
	for _, clusterRole := range clusterRoles {
		rules := clusterRole.Rules

		for _, rule := range rules {
			groups := rule.APIGroups
			sort.Slice(groups, func(i, j int) bool {
				return groups[i] < groups[j]
			})

			resources := rule.Resources
			sort.Slice(resources, func(i, j int) bool {
				return resources[i] < resources[j]
			})

			verbs := rule.Verbs
			sort.Slice(verbs, func(i, j int) bool {
				return verbs[i] < verbs[j]
			})
		}

		sort.Slice(rules, func(i, j int) bool {
			ruleI := rules[i]
			stringI := fmt.Sprintf("%v%v%v", ruleI.APIGroups, ruleI.Resources, ruleI.Verbs)

			ruleJ := rules[j]
			stringJ := fmt.Sprintf("%v%v%v", ruleJ.APIGroups, ruleJ.Resources, ruleJ.Verbs)

			return stringI < stringJ
		})
	}

	sort.Slice(clusterRoles, func(i, j int) bool {
		return clusterRoles[i].GetName() < clusterRoles[j].GetName()
	})
}
