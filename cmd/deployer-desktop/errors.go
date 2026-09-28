package main

import (
	"fmt"
	"strings"
)

type errorCategory string

const (
	categoryValidation errorCategory = "validation"
	categoryConnection errorCategory = "connection"
	categoryInspection errorCategory = "inspection"
	categoryBuild      errorCategory = "build"
	categoryDeployment errorCategory = "deployment"
	categoryUnexpected errorCategory = "unexpected"
)

// categorized formats errors as "<category>: <message>". Deliberate
// compromise: a fully structured {ok, category, message} return type
// on every bound method would be "more correct" but means rewriting
// every call site on both sides of the bridge for a distinction only
// a few error paths actually need. This convention lets every method
// keep returning a plain `error` (which Wails already knows how to
// surface to JS) while the frontend recovers a category by splitting
// on the first ": ".
func categorized(cat errorCategory, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %s", cat, err.Error())
}

func validationError(msg string) error {
	return fmt.Errorf("%s: %s", categoryValidation, msg)
}

// classifyDeployError inspects deploy.Run's own error prefixes
// ("preparing workspace:", "docker build:", "docker run:", "uploading
// to", "connecting to") to assign a category. deploy.Run itself has
// no notion of "categories" — this is the Wails-boundary adaptation
// Phase 5 asked for, rather than teaching the engine about desktop
// concepts it shouldn't need to know.
func classifyDeployError(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "docker not available"),
		strings.Contains(msg, "uploading to"),
		strings.Contains(msg, "connecting to"):
		return categorized(categoryConnection, err)
	case strings.Contains(msg, "preparing workspace"),
		strings.Contains(msg, "docker build"):
		return categorized(categoryBuild, err)
	case strings.Contains(msg, "docker run"):
		return categorized(categoryDeployment, err)
	default:
		return categorized(categoryUnexpected, err)
	}
}