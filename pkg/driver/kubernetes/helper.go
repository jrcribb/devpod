package kubernetes

import (
	"fmt"
	"os"
	"strings"

	"github.com/loft-sh/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes/scheme"
)

const (
	limitsPrefix   = "limits."
	requestsPrefix = "requests."
)

func parseResources(resourceString string, log log.Logger) corev1.ResourceRequirements {
	if resourceString == "" {
		return corev1.ResourceRequirements{}
	}

	resourcesSplitted := strings.Split(resourceString, ",")
	requests := corev1.ResourceList{}
	limits := corev1.ResourceList{}
	for _, resourceName := range resourcesSplitted {
		resourceName = strings.TrimSpace(resourceName)

		// requests
		if strings.HasPrefix(resourceName, requestsPrefix) {
			strippedResource := strings.TrimPrefix(resourceName, requestsPrefix)
			name, quantity, err := parseResource(strippedResource)
			if err != nil {
				log.Error(err.Error())
				continue
			}

			requests[corev1.ResourceName(name)] = quantity
		}

		// limits
		if strings.HasPrefix(resourceName, limitsPrefix) {
			strippedResource := strings.TrimPrefix(resourceName, limitsPrefix)
			name, quantity, err := parseResource(strippedResource)
			if err != nil {
				log.Error(err.Error())
				continue
			}

			limits[corev1.ResourceName(name)] = quantity
		}
	}

	return corev1.ResourceRequirements{
		Limits:   limits,
		Requests: requests,
	}
}

func getPodTemplate(manifestFilePath string) (*corev1.Pod, error) {
	body, err := os.ReadFile(manifestFilePath)
	if err != nil {
		return nil, err
	}
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(body, nil, nil)
	if err != nil {
		return nil, err
	}
	switch resource := obj.(type) {
	case *corev1.Pod:
		return resource, nil
	}
	return nil, fmt.Errorf("no pod manifest found in file %s", manifestFilePath)
}

func parseLabels(str string) (map[string]string, error) {
	if str == "" {
		return nil, nil
	}

	labels := strings.Split(str, ",")
	retMap := map[string]string{}
	for _, label := range labels {
		label = strings.TrimSpace(label)
		splitted := strings.SplitN(label, "=", 2)
		if len(splitted) != 2 {
			return nil, fmt.Errorf("invalid label '%s', expected format label=value", label)
		}

		retMap[splitted[0]] = splitted[1]
	}

	return retMap, nil
}

func parseResource(resourceName string) (string, resource.Quantity, error) {
	splittedResource := strings.SplitN(resourceName, "=", 2)
	if len(splittedResource) != 2 {
		return "", resource.Quantity{}, fmt.Errorf("error parsing resource %s: expected form resource=quantity", resourceName)
	}

	quantity, err := resource.ParseQuantity(splittedResource[1])
	if err != nil {
		return "", resource.Quantity{}, fmt.Errorf("error parsing resource %s: %w", resourceName, err)
	}

	return splittedResource[0], quantity, nil
}
