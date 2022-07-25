#!/usr/bin/env bats
@test "pcm-catalogs empty get collection" {
	run epcc get pcm-catalogs
	[ $status -eq 0 ]
}

@test "pcm-catalogs delete-all support" {
	run epcc delete-all pcm-catalogs
	[ $status -eq 0 ]
}


