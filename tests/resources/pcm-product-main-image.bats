#!/usr/bin/env bats
@test "pcm-product-main-image empty get collection" {
	run epcc get pcm-product-main-image
	[ $status -eq 0 ]
}

