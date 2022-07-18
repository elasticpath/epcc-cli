#!/usr/bin/env bats
@test "pcm-variation-options empty get collection" {
	run epcc get pcm-variation-options
	[ $status -eq 0 ]
}

