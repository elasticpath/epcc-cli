#!/usr/bin/env bats
@test "pcm-hierarchies empty get collection" {
	run epcc get pcm-hierarchies
	[ $status -eq 0 ]
}

@test "pcm-hierarchies delete-all support" {
	run epcc delete-all pcm-hierarchies
	[ $status -eq 0 ]
}


