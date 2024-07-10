package tmpl

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/itchyny/gojq"
)

type JQTemplate interface {
	Execute(query string, data any) (any, error)
}

var _ JQTemplate = (*jqTemplate)(nil)

func New(leftDelim, rightDelim string) (JQTemplate, error) {
	pattern := fmt.Sprintf("^%s\\s+(.*)%s",
		regexp.QuoteMeta(leftDelim),
		regexp.QuoteMeta(rightDelim))

	re, err := regexp.Compile(pattern) // `^\{\{\s+(.*)\}\}$`)
	if err != nil {
		return nil, err
	}

	return &jqTemplate{
		re:      re,
		unquote: true,
	}, nil
}

type jqTemplate struct {
	re      *regexp.Regexp
	unquote bool
}

func (t *jqTemplate) Execute(q string, data any) (any, error) {
	q, ok := t.acceptQuery(q)
	if !ok {
		return q, nil
	}

	// enc := newEncoder(false, 0)

	query, err := gojq.Parse(q)
	if err != nil {
		return "", err
	}

	iter := query.Run(data) // or query.RunWithContext
	// only one result is expected
	v, ok := iter.Next()
	if !ok {
		return "", nil
	}
	if err, ok := v.(error); ok {
		return "", err
	}
	return v, nil
}

func (t *jqTemplate) acceptQuery(q string) (string, bool) {
	if !t.re.MatchString(q) {
		return q, false
	}

	res := t.re.FindAllStringSubmatch(q, -1)
	if len(res) == 0 || len(res[0]) == 0 {
		return q, false
	}

	return strings.TrimSpace(res[0][1]), true
}
