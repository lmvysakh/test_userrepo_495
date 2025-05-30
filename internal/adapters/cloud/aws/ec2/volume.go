package ec2

import (
	"fmt"

	ec2api "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/aquasecurity/trivy-aws/pkg/concurrency"
	"github.com/aquasecurity/trivy/pkg/iac/providers/aws/ec2"
	trivyTypes "github.com/aquasecurity/trivy/pkg/iac/types"
)

func (a *adapter) getVolumes() ([]ec2.Volume, error) {

	a.Tracker().SetServiceLabel("Discovering volumes...")

	var input ec2api.DescribeVolumesInput

	var apiVolumes []types.Volume
	for {
		output, err := a.client.DescribeVolumes(a.Context(), &input)
		if err != nil {
			return nil, err
		}
		apiVolumes = append(apiVolumes, output.Volumes...)
		a.Tracker().SetTotalResources(len(apiVolumes))
		if output.NextToken == nil {
			break
		}
		input.NextToken = output.NextToken
	}

	a.Tracker().SetServiceLabel("Adapting volumes...")
	return concurrency.Adapt(apiVolumes, a.RootAdapter, a.adaptVolume), nil
}

func (a *adapter) adaptVolume(volume types.Volume) (*ec2.Volume, error) {

	metadata := a.CreateMetadata(fmt.Sprintf("volume/%s", *volume.VolumeId))

	encrypted := volume.Encrypted != nil && *volume.Encrypted
	var kmsKeyID string
	if volume.KmsKeyId != nil {
		kmsKeyID = *volume.KmsKeyId
	}

	return &ec2.Volume{
		Metadata: metadata,
		Encryption: ec2.Encryption{
			Metadata: metadata,
			Enabled:  trivyTypes.Bool(encrypted, metadata),
			KMSKeyID: trivyTypes.String(kmsKeyID, metadata),
		},
	}, nil
}
