package ssm

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// GetSecrets fetches parameters by prefix
func GetSecrets(ctx context.Context, client *ssm.Client, path string, recursive bool, trimKeyName bool) (map[string]string, error) {
	secrets := map[string]string{}

	paginator := ssm.NewGetParametersByPathPaginator(client, &ssm.GetParametersByPathInput{
		Path:           aws.String(path),
		Recursive:      aws.Bool(recursive),
		WithDecryption: aws.Bool(true),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, p := range page.Parameters {
			name := aws.ToString(p.Name)
			value := aws.ToString(p.Value)

			if trimKeyName {
				name = name[strings.LastIndex(name, "/")+1:]
			} else {
				name = name
			}
			secrets[name] = value
		}
	}

	return secrets, nil
}
