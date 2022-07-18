#!/usr/bin/env bats
@test "pcm-hierarchies empty get collection" {
	run epcc get pcm-hierarchies
	[ $status -eq 0 ]
}

