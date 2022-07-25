#!/usr/bin/env bats
@test "pcm-product-main-image empty get collection" {
	run epcc get pcm-product-main-image
	[ $status -eq 0 ]
}

@test "pcm-product-main-image delete-all support" {
	run epcc delete-all pcm-product-main-image
	[ $status -eq 0 ]
}


