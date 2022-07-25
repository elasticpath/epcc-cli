#!/usr/bin/env bats
@test "pcm-pricebooks empty get collection" {
	run epcc get pcm-pricebooks
	[ $status -eq 0 ]
}

@test "pcm-pricebooks delete-all support" {
	run epcc delete-all pcm-pricebooks
	[ $status -eq 0 ]
}


