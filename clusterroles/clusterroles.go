package main

import (
	"fmt"
	"io"
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
	separatorGroupVersion = "/"
	separatorPanic        = ": "
	separatorYaml         = "---\n"

	VerbGet   = "get"
	VerbList  = "list"
	VerbWatch = "watch"

	CoreGroupName      = ""
	secretResourceName = "secrets"

	NamespacedReadOnlyRoleName    = "namespaced-ro"
	NamespacedReadWriteRoleName   = "namespaced-rw"
	UnnamespacedReadOnlyRoleName  = "unnamespaced-ro"
	UnnamespacedReadWriteRoleName = "unnamespaced-rw"
)

var (
	readOnlyVerbs = []string{
		VerbGet,
		VerbList,
		VerbWatch,
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

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Panic(filePath, separatorPanic, err)
	}

	apiResourceLists, err := getApiResourceLists(data)
	if err != nil {
		log.Panic(filePath, separatorPanic, err)
	}

	if err := GenerateManifests(apiResourceLists, os.Stdout); err != nil {
		log.Panic(filePath, separatorPanic, err)
	}
}

func getApiResourceLists(data []byte) ([]*metav1.APIResourceList, error) {
	clusterRoles := ClusterRoles{
		KubeConfig: ClusterRolesKubeConfig{
			LoadingRules: clientcmd.NewDefaultClientConfigLoadingRules(),
			Overrides:    &clientcmd.ConfigOverrides{},
		},
	}
	if err := yaml.Unmarshal(data, &clusterRoles); err != nil {
		return nil, err
	}

	deferredLoadingClientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clusterRoles.KubeConfig.LoadingRules,
		clusterRoles.KubeConfig.Overrides,
	)
	clientConfig, err := deferredLoadingClientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(clientConfig)
	if err != nil {
		return nil, err
	}
	_, apiResourceLists, err := discoveryClient.ServerGroupsAndResources()

	return apiResourceLists, err
}

func GenerateManifests(apiResourceLists []*metav1.APIResourceList, out io.Writer) error {
	groupIndex, err := makeGroupIndex(apiResourceLists)
	if err != nil {
		return err
	}
	clusterRoles, err := makeClusterRoles(groupIndex)
	if err != nil {
		return err
	}

	for _, clusterRole := range clusterRoles {
		if _, err := out.Write([]byte(separatorYaml)); err != nil {
			return err
		}

		b, err := yaml.Marshal(clusterRole)
		if err != nil {
			return err
		}
		if _, err := out.Write(b); err != nil {
			return err
		}
	}

	return nil
}

type Namespaced bool
type ResourceIndex map[string]Namespaced
type GroupIndex map[string]ResourceIndex

func makeGroupIndex(apiResourceLists []*metav1.APIResourceList) (GroupIndex, error) {
	groupIndex := make(GroupIndex)
	for _, resourceList := range apiResourceLists {
		groupVersion := resourceList.GroupVersion

		var groupName string
		if indexSeparator := strings.Index(groupVersion, separatorGroupVersion); indexSeparator != -1 {
			groupName = groupVersion[:indexSeparator]
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

func makeClusterRoles(groupIndex GroupIndex) ([]rbacv1.ClusterRole, error) {
	var clusterRoles []rbacv1.ClusterRole

	namespacedRoles, err := makeNamespacedClusterRoles(groupIndex)
	if err != nil {
		return nil, err
	}
	clusterRoles = append(clusterRoles, namespacedRoles...)

	unnamespacedRoles, err := makeUnnamespacedClusterRoles(groupIndex)
	if err != nil {
		return nil, err
	}
	clusterRoles = append(clusterRoles, unnamespacedRoles...)

	canonicalizeClusterRoles(clusterRoles)

	return clusterRoles, nil
}

func makeNamespacedClusterRoles(groupIndex GroupIndex) ([]rbacv1.ClusterRole, error) {
	coreRule := rbacv1.PolicyRule{
		APIGroups: []string{
			CoreGroupName,
		},
		Verbs: readOnlyVerbs,
	}

	othersRule := rbacv1.PolicyRule{
		Resources: []string{
			rbacv1.ResourceAll,
		},
		Verbs: readOnlyVerbs,
	}

	for group, resources := range groupIndex {
		if group != CoreGroupName {
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
		{
			TypeMeta: typeMeta,
			ObjectMeta: metav1.ObjectMeta{
				Name: NamespacedReadOnlyRoleName,
			},
			Rules: []rbacv1.PolicyRule{
				coreRule,
				othersRule,
			},
		},
		{
			TypeMeta: typeMeta,
			ObjectMeta: metav1.ObjectMeta{
				Name: NamespacedReadWriteRoleName,
			},
			Rules: []rbacv1.PolicyRule{
				{
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

func makeUnnamespacedClusterRoles(groupIndex GroupIndex) ([]rbacv1.ClusterRole, error) {
	var readOnlyRules []rbacv1.PolicyRule
	var readWriteRules []rbacv1.PolicyRule

	for group, resources := range groupIndex {
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
		{
			TypeMeta: typeMeta,
			ObjectMeta: metav1.ObjectMeta{
				Name: UnnamespacedReadOnlyRoleName,
			},
			Rules: readOnlyRules,
		},
		{
			TypeMeta: typeMeta,
			ObjectMeta: metav1.ObjectMeta{
				Name: UnnamespacedReadWriteRoleName,
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
