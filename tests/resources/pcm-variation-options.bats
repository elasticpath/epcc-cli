#!/usr/bin/env bats
@test "pcm-variation-options empty get collection" {
	run epcc get pcm-variation-options
	[ $status -eq 0 ]
}

@test "pcm-variation-options delete-all support" {
	run epcc delete-all pcm-variation-options
	[ $status -eq 0 ]
}


