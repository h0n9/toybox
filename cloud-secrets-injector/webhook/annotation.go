package webhook

import (
	"strconv"
	"strings"
)

type Annotations map[string]string

const (
	AnnotationPrefix = "cloud-secrets-injector.h0n9.postie.chat"
)

var annotationsAvailable = map[string]bool{
	"provider": true,
	"key-id":   true,
	"template": true,
	"output":   true,
	"injected": true,
}

func ParseAndCheckAnnotations(input Annotations) Annotations {
	output := map[string]string{}
	for key, value := range input {
		subPath := strings.TrimPrefix(key, AnnotationPrefix+"/")
		if subPath == key {
			continue
		}
		if _, exist := annotationsAvailable[subPath]; !exist {
			continue
		}
		output[subPath] = value
	}
	return output
}

func (a Annotations) IsInected() bool {
	value, exist := a["injected"]
	if !exist {
		return false
	}
	injected, err := strconv.ParseBool(value)
	if err != nil {
		return false
	}
	return injected
}
