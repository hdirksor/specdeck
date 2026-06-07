package build_test

import (
	"testing"

	"github.com/hdickson/specdeck/internal/build"
)

func TestValidate_Noop(t *testing.T) {
	containers := []build.Container{
		{Title: "Screen", Specs: build.Specs{"color": "blue"}},
	}
	if err := build.Validate(containers); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}
