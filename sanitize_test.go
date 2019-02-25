package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMakeUrlSlug(t *testing.T) {
	testSet := map[string]string{
		"my-featasdasdaaure-0009-my-app12": "1390äö-m_#y-featasdasdaaure-0009-my-app12-",
		"m":                                "-m-",
		"m-m":                              "-2-m-m-",
		"feature-foo-999-automate-staging-deploy-ffffff-app123456789000": "feature-FOO-999-automate-staging-deploy-ffffff-app123456789000-foo",
	}

	assert := assert.New(t)

	for expected, value := range testSet {
		actual := MakeUrlSlug(value, DNS_MAX_LENGTH)

		assert.Equal(expected, actual, "The two words should be the same.")
	}
}
