package kubernetes

import (
	"strconv"

	apps "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func flattenDeploymentSpec(in apps.DeploymentSpec) ([]interface{}, error) {
	att := make(map[string]interface{})
	att["min_ready_seconds"] = in.MinReadySeconds

	if in.Replicas != nil {
		att["replicas"] = *in.Replicas
	}

	att["selector"] = in.Selector.MatchLabels
	att["strategy"] = flattenDeploymentStrategy(in.Strategy)
	podSpec, err := flattenPodSpec(in.Template.Spec)
	if err != nil {
		return nil, err
	}
	att["template"] = podSpec

	return []interface{}{att}, nil
}

func flattenDeploymentStrategy(in apps.DeploymentStrategy) []interface{} {
	att := make(map[string]interface{})
	if in.Type != "" {
		att["type"] = in.Type
	}
	if in.RollingUpdate != nil {
		att["rollingUpdate"] = flattenDeploymentStrategyRollingUpdate(in.RollingUpdate)
	}
	return []interface{}{att}
}

func flattenDeploymentStrategyRollingUpdate(in *apps.RollingUpdateDeployment) []interface{} {
	att := make(map[string]interface{})
	if in.MaxSurge != nil {
		att["maxSurge"] = in.MaxSurge.String()
	}
	if in.MaxUnavailable != nil {
		att["maxUnavailable"] = in.MaxUnavailable.String()
	}
	return []interface{}{att}
}

func expandDeploymentSpec(deployment []interface{}) (apps.DeploymentSpec, error) {
	obj := apps.DeploymentSpec{}
	if len(deployment) == 0 || deployment[0] == nil {
		return obj, nil
	}
	in := deployment[0].(map[string]interface{})
	obj.MinReadySeconds = int32(in["min_ready_seconds"].(int))
	obj.Replicas = ptrToInt32(int32(in["replicas"].(int)))
	obj.Selector = &metav1.LabelSelector{
		MatchLabels: expandStringMap(in["selector"].(map[string]interface{})),
	}
	obj.Strategy = expandDeploymentStrategy(in["strategy"].([]interface{}))
	podSpec, err := expandPodSpec(in["template"].([]interface{}))
	if err != nil {
		return obj, err
	}
	obj.Template = v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: obj.Selector.MatchLabels,
		},
		Spec: podSpec,
	}

	return obj, nil
}

func expandDeploymentStrategy(p []interface{}) apps.DeploymentStrategy {
	obj := apps.DeploymentStrategy{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["type"]; ok {
		obj.Type = apps.DeploymentStrategyType(v.(string))
	}
	if v, ok := in["rollingUpdate"]; ok {
		obj.RollingUpdate = expandRollingUpdateDeployment(v.([]interface{}))
	}
	return obj
}

func expandRollingUpdateDeployment(p []interface{}) *apps.RollingUpdateDeployment {
	obj := apps.RollingUpdateDeployment{}
	if len(p) == 0 || p[0] == nil {
		return &obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["maxSurge"]; ok {
		obj.MaxSurge = expandRollingUpdateDeploymentIntOrString(v.(string))
	}
	if v, ok := in["maxUnavailable"]; ok {
		obj.MaxUnavailable = expandRollingUpdateDeploymentIntOrString(v.(string))
	}
	return &obj
}

func expandRollingUpdateDeploymentIntOrString(v string) *intstr.IntOrString {
	i, err := strconv.Atoi(v)
	if err != nil {
		return &intstr.IntOrString{
			Type:   intstr.String,
			StrVal: v,
		}
	}
	return &intstr.IntOrString{
		Type:   intstr.Int,
		IntVal: int32(i),
	}
}
