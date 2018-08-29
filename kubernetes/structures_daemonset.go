package kubernetes

import (
	"strconv"

	apps "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func flattenDaemonSetSpec(in apps.DaemonSetSpec) ([]interface{}, error) {
	att := make(map[string]interface{})
	att["min_ready_seconds"] = in.MinReadySeconds

	att["selector"] = in.Selector.MatchLabels
	att["strategy"] = flattenDaemonSetStrategy(in.UpdateStrategy)
	podSpec, err := flattenPodSpec(in.Template.Spec)
	if err != nil {
		return nil, err
	}
	att["template"] = podSpec

	return []interface{}{att}, nil
}

func flattenDaemonSetStrategy(in apps.DaemonSetUpdateStrategy) []interface{} {
	att := make(map[string]interface{})
	if in.Type != "" {
		att["type"] = in.Type
	}
	if in.RollingUpdate != nil {
		att["rollingUpdate"] = flattenDaemonSetStrategyRollingUpdate(in.RollingUpdate)
	}
	return []interface{}{att}
}

func flattenDaemonSetStrategyRollingUpdate(in *apps.RollingUpdateDaemonSet) []interface{} {
	att := make(map[string]interface{})
	if in.MaxUnavailable != nil {
		att["maxUnavailable"] = in.MaxUnavailable.String()
	}
	return []interface{}{att}
}

func expandDaemonSetSpec(deployment []interface{}) (apps.DaemonSetSpec, error) {
	obj := apps.DaemonSetSpec{}
	if len(deployment) == 0 || deployment[0] == nil {
		return obj, nil
	}
	in := deployment[0].(map[string]interface{})
	obj.MinReadySeconds = int32(in["min_ready_seconds"].(int))
	if v, ok := in["selector"]; ok {
		obj.Selector = &metav1.LabelSelector{
			MatchLabels: expandStringMap(v.(map[string]interface{})),
		}
	}
	obj.UpdateStrategy = expandDaemonSetStrategy(in["strategy"].([]interface{}))
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

func expandDaemonSetStrategy(p []interface{}) apps.DaemonSetUpdateStrategy {
	obj := apps.DaemonSetUpdateStrategy{}

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["type"]; ok {
		obj.Type = apps.DaemonSetUpdateStrategyType(v.(string))
	}
	if v, ok := in["rollingUpdate"]; ok {
		obj.RollingUpdate = expandRollingUpdateDaemonSet(v.([]interface{}))
	}
	return obj
}

func expandRollingUpdateDaemonSet(p []interface{}) *apps.RollingUpdateDaemonSet {
	obj := apps.RollingUpdateDaemonSet{}
	if len(p) == 0 || p[0] == nil {
		return &obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["maxUnavailable"]; ok {
		obj.MaxUnavailable = expandRollingUpdateDaemonSetIntOrString(v.(string))
	}
	return &obj
}

func expandRollingUpdateDaemonSetIntOrString(v string) *intstr.IntOrString {
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
