package main

import (
	"github.com/1204244136/MDA/agent/go-service/common/myaction"
	"github.com/1204244136/MDA/agent/go-service/common/myreco"
	"github.com/1204244136/MDA/agent/go-service/pkg/resource"
	"github.com/1204244136/MDA/agent/go-service/taskersink/aspectratio"
	"github.com/1204244136/MDA/agent/go-service/taskersink/hdrcheck"
	"github.com/rs/zerolog/log"
)

func registerAll() {
	// Resource Sink
	resource.EnsureResourcePathSink()

	// Pre-Check Custom
	aspectratio.Register()
	hdrcheck.Register()

	// Custom Actions
	myaction.Register()

	// Custom Recognitions
	myreco.Register()

	log.Info().
		Msg("All custom components and sinks registered successfully")
}
