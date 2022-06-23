package webhook

import "strings"

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

func ParseAndCheckAnnotations(input map[string]string) map[string]string {
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
