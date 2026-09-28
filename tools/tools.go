//go:build generate

package tools

import (
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
)

//go:generate terraform fmt -recursive ../examples/provider/ ../examples/resources/ ../examples/data-sources/
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-dir .. -provider-name postgresql -tf-version 1.15.4
