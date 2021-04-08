package k8sclient

import (
	"fmt"
	"strings"

	"github.com/golang/glog"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	"k8s.io/client-go/rest"
)

const HPAKind = "horizontalpodautoscaler"

// ~= updateReplicasAppsV1()
func (k *k8sClient) updateReplicasHPA(expReplicas int32) (prevRelicas int32, err error) {
	if k.target.kind != HPAKind {
		return 0, fmt.Errorf("unexpected target kind: %v", k.target.kind)
	}

	req, err := requestForHPA(k.clientset.AutoscalingV1().RESTClient().Get(), k.target)
	if err != nil {
		return 0, err
	}

	hpa := &autoscalingv1.HorizontalPodAutoscaler{}
	if err = req.Do().Into(hpa); err != nil {
		return 0, err
	}

	prevRelicas = *hpa.Spec.MinReplicas
	if expReplicas != prevRelicas {
		glog.V(0).Infof("Cluster status: SchedulableNodes[%v], SchedulableCores[%v]", k.clusterStatus.SchedulableNodes, k.clusterStatus.SchedulableCores)
		glog.V(0).Infof("MinReplicas is not as expected : updating minReplicas from %d to %d", prevRelicas, expReplicas)
		*hpa.Spec.MinReplicas = expReplicas
		req, err = requestForHPA(k.clientset.AutoscalingV1().RESTClient().Put(), k.target)
		if err != nil {
			return 0, err
		}
		if err = req.Body(hpa).Do().Error(); err != nil {
			return 0, err
		}
	}

	return prevRelicas, nil
}

// ~= requestForTarget
func requestForHPA(req *rest.Request, target *scaleTarget) (*rest.Request, error) {
	var absPath, resource string
	switch strings.ToLower(target.kind) {
	case "horizontalpodautoscaler":
		absPath = "/apis/autoscaling/v1"
		resource = "horizontalpodautoscalers"
	default:
		return nil, fmt.Errorf("unsupported target kind: %v", target.kind)
	}

	return req.AbsPath(absPath).Namespace(target.namespace).Resource(resource).Name(target.name), nil
}
