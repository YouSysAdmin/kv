package amazon

import (
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

func New(ctx context.Context, profile, region, accessKeyID, accessKeySecret string) (aws.Config, error) {
	var cfg aws.Config
	var err error

	if accessKeyID != "" && accessKeySecret != "" {
		creds := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(
			accessKeyID,
			accessKeySecret,
			"",
		))
		cfg, err = config.LoadDefaultConfig(ctx, config.WithCredentialsProvider(creds), config.WithRegion(region))
		if err != nil {
			if strings.Contains(err.Error(), "missing region") || strings.Contains(err.Error(), "MissingRegion") {
				return cfg, errors.New("AWS region is not configured. Use --region or configure it in your AWS profile")
			}
			return cfg, err
		}
	} else {
		var opts []func(*config.LoadOptions) error
		if profile != "" {
			opts = append(opts, config.WithSharedConfigProfile(profile))
		}
		if region != "" {
			opts = append(opts, config.WithRegion(region))
		}
		cfg, err = config.LoadDefaultConfig(ctx, opts...)
		if err != nil {
			if strings.Contains(err.Error(), "missing region") || strings.Contains(err.Error(), "MissingRegion") {
				return cfg, errors.New("AWS region is not configured. Use --region or configure it in your AWS profile")
			}
			return cfg, err
		}
	}

	return cfg, nil
}
