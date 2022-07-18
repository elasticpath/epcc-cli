#!/usr/bin/env bats
@test "pcm-catalogs empty get collection" {
	run epcc get pcm-catalogs
	[ $status -eq 0 ]
}

