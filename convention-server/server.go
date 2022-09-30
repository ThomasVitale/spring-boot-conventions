package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"

	"github.com/arktonix/spring-boot-conventions/convention-server/resources"
	"github.com/vmware-tanzu/cartographer-conventions/webhook"
)

func addSpringBootConventions(template *corev1.PodTemplateSpec, images []webhook.ImageConfig) ([]string, error) {
	imageMap := make(map[string]webhook.ImageConfig)
	for _, config := range images {
		imageMap[config.Image] = config
	}

	var appliedConventions []string
	for i := range template.Spec.Containers {
		container := &template.Spec.Containers[i]
		image, ok := imageMap[container.Image]
		if !ok {
			// Skip containers without metadata, this may be a container without an image
			continue
		}
		dependencyMetadata := resources.NewDependenciesBOM(image.BOMs)
		applicationProperties := resources.SpringApplicationProperties{}
		applicationProperties.FromContainer(container)

		ctx := context.Background()
		ctx = resources.StashSpringApplicationProperties(ctx, applicationProperties)
		ctx = resources.StashDependenciesBOM(ctx, &dependencyMetadata)
		for _, o := range resources.SpringBootConventions {
			// Need to continue refining what metadata is passed to the conventions.
			if !o.IsApplicable(ctx, imageMap) {
				continue
			}
			appliedConventions = append(appliedConventions, o.GetId())
			if err := o.ApplyConvention(ctx, template, i, imageMap); err != nil {
				return nil, err
			}
		}
		applicationProperties.ToContainer(container)
	}
	return appliedConventions, nil
}

func main() {
	ctx := context.Background()
	port := os.Getenv("PORT")
	if port == "" {
		port = "9000"
	}

	zapLog, err := zap.NewProductionConfig().Build()
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}

	logger := zapr.NewLogger(zapLog)
	ctx = logr.NewContext(ctx, logger)

	http.HandleFunc("/", webhook.ConventionHandler(ctx, addSpringBootConventions))
	log.Fatal(webhook.NewConventionServer(ctx, fmt.Sprintf(":%s", port)))
}
