#!/usr/bin/env bats
@test "pcm-products empty get collection" {
	run epcc get pcm-products
	[ $status -eq 0 ]
}

@test "pcm-products delete-all support" {
	run epcc delete-all pcm-products
	[ $status -eq 0 ]
}


