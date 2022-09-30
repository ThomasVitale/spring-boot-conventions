package resources

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

var _ Convention = (*SpringBootServiceIntent)(nil)

type SpringBootServiceIntent struct {
	Id           string
	LabelName    string
	Dependencies []string
}

func (o *SpringBootServiceIntent) GetId() string {
	return o.Id
}

func (o *SpringBootServiceIntent) IsApplicable(ctx context.Context, metadata ImageMetadata) bool {
	deps := GetDependenciesBOM(ctx)
	return deps.HasDependency(o.Dependencies...)
}

func (o *SpringBootServiceIntent) ApplyConvention(ctx context.Context, target *corev1.PodTemplateSpec, containerIdx int, metadata ImageMetadata) error {
	deps := GetDependenciesBOM(ctx)
	for _, d := range o.Dependencies {
		if dbom := deps.Dependency(d); dbom != nil {
			setLabel(target, o.LabelName, target.Spec.Containers[containerIdx].Name)
			setAnnotation(target, o.LabelName, fmt.Sprintf("%s/%s", dbom.Name, dbom.Version))
			break
		}
	}
	return nil
}
