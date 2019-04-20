package main

import (
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func TestConfigSuite(t *testing.T) {
	id := "430123199001111234"
	reg := regexp.MustCompile(`(^[1-9]\d{5}(18|19|([23]\d))\d{2}((0[1-9])|(10|11|12))(([0-2][1-9])|10|20|30|31)\d{3}[0-9Xx]$)|(^[1-9]\d{5}\d{2}((0[1-9])|(10|11|12))(([0-2][1-9])|10|20|30|31)\d{3}$)`)
	ok := reg.MatchString(id)
	assert.True(t, ok)
}
