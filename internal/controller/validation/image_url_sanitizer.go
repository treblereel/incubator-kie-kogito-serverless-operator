// Copyright 2024 Apache Software Foundation (ASF)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package validation

import (
	"context"
	"fmt"
	"net"

	operatorapi "github.com/apache/incubator-kie-kogito-serverless-operator/api/v1alpha08"
	"github.com/google/go-containerregistry/pkg/name"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type imageUrlSanitizer struct{}

func (v *imageUrlSanitizer) Validate(ctx context.Context, client client.Client, sonataflow *operatorapi.SonataFlow, req ctrl.Request) error {
	isInKindRegistry, kindRegistryUrl, err := ImageStoredInKindRegistry(ctx, sonataflow.Spec.PodTemplate.Container.Image)
	if err != nil {
		return err
	}
	if isInKindRegistry {
		ref, err := name.ParseReference(sonataflow.Spec.PodTemplate.Container.Image)
		if err != nil {
			return fmt.Errorf("Unable to parse image uri: %w", err)
		}
		// host or ip address and port
		hostAndPortName := ref.Context().RegistryStr()

		host, _, err := net.SplitHostPort(hostAndPortName)
		if err != nil {
			return fmt.Errorf("Unable to parse image uri: %w", err)
		}
		// check if host is not ip address,
		if net.ParseIP(host) == nil {
			// take part after port
			var result string
			// if tag isn't present, it's latest
			if tag, ok := ref.(name.Tag); ok {
				result = fmt.Sprintf("%s/%s:%s", kindRegistryUrl, ref.Context().RepositoryStr(), tag.TagStr())
			} else {
				result = fmt.Sprintf("%s/%s", kindRegistryUrl, ref.Context().RepositoryStr())
			}
			sonataflow.Spec.PodTemplate.Container.Image = result
			if err := client.Update(ctx, sonataflow); err != nil {
				return fmt.Errorf("failed to update SonataFlow object: %w", err)
			}
		}
	}

	return nil
}

func NewImageUrlSanitizer() Validator {
	return &imageUrlSanitizer{}
}
