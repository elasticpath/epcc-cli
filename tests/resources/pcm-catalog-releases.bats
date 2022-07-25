#!/usr/bin/env bats
@test "pcm-catalog-releases empty get collection" {
	run epcc get pcm-catalog-releases
	[ $status -eq 0 ]
}

@test "pcm-catalog-releases delete-all support" {
	run epcc delete-all pcm-catalog-releases
	[ $status -eq 0 ]
}


