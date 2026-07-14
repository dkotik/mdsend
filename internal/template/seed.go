package template

import (
	"fmt"
	ttemplate "text/template"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/cespare/xxhash/v2"
	"github.com/dkotik/mdsend"
)

const DefaultSeedTemplate = "{{ .Frontmatter.subject }}||{{ .Content }}"

func newSeedKey(ctx Context, l mdsend.Letter) (string, error) {
	seedString, err := l.GetSeed()
	if err != nil {
		return "", fmt.Errorf("invalid seed: %w", err)
	}
	if seedString == "" {
		seedString = DefaultSeedTemplate
	}
	seedTemplate, err := ttemplate.New("").Parse(seedString)
	if err != nil {
		return "", fmt.Errorf("invalid seed template: %w", err)
	}
	h := xxhash.New()
	if err = seedTemplate.Execute(h, ctx); err != nil {
		return "", fmt.Errorf("unable to execute seed template: %w", err)
	}

	return base58.Encode(h.Sum(nil)), nil
}
