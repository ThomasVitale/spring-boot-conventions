package resources

import (
	"context"
	"fmt"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

type SpringApplicationProperties map[string]string

func (p SpringApplicationProperties) Default(key string, defaultValue string) string {
	if _, ok := p[key]; !ok {
		p[key] = defaultValue
	}
	return p[key]
}

func (p SpringApplicationProperties) FromContainer(c *corev1.Container) {
	javaOpts := findEnvVar(*c, "JAVA_TOOL_OPTIONS")
	if javaOpts == nil {
		return
	}
	keep := []string{}
	for _, c := range strings.Split(javaOpts.Value, " ") {
		if !strings.HasPrefix(c, "-D") || !strings.Contains(c, "=") {
			keep = append(keep, c)
			continue
		}
		kv := strings.SplitN(c[2:], "=", 2)
		p[kv[0]] = kv[1]
	}
	// Remove opts as they will be added back after the conventions are applied.
	javaOpts.Value = strings.Join(keep, " ")
}

func (p SpringApplicationProperties) ToContainer(c *corev1.Container) {
	properties := []string{}
	propertyKeys := []string{}
	for key := range p {
		propertyKeys = append(propertyKeys, key)
	}
	sort.Strings(propertyKeys)
	for _, key := range propertyKeys {
		properties = append(properties, fmt.Sprintf("-D%s=%s", key, p[key]))
	}

	// Set application properties on JAVA_TOOL_OPTIONS.
	javaOpts := findEnvVar(*c, "JAVA_TOOL_OPTIONS")
	if javaOpts != nil {
		javaOpts.Value = strings.TrimSpace(fmt.Sprintf("%s %s", javaOpts.Value, strings.Join(properties, " ")))
	} else {
		c.Env = append(c.Env, corev1.EnvVar{
			Name:  "JAVA_TOOL_OPTIONS",
			Value: strings.Join(properties, " "),
		})
	}
}

type springApplicationPropertiesKey struct{}

func StashSpringApplicationProperties(ctx context.Context, props SpringApplicationProperties) context.Context {
	return context.WithValue(ctx, springApplicationPropertiesKey{}, props)
}

func GetSpringApplicationProperties(ctx context.Context) SpringApplicationProperties {
	value := ctx.Value(springApplicationPropertiesKey{})
	if props, ok := value.(SpringApplicationProperties); ok {
		return props
	}

	return SpringApplicationProperties{}
}
