#!/usr/bin/env bats
@test "pcm-variations empty get collection" {
	run epcc get pcm-variations
	[ $status -eq 0 ]
}

@test "pcm-variations delete-all support" {
	run epcc delete-all pcm-variations
	[ $status -eq 0 ]
}


