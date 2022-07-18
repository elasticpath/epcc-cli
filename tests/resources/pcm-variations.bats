#!/usr/bin/env bats
@test "pcm-variations empty get collection" {
	run epcc get pcm-variations
	[ $status -eq 0 ]
}

