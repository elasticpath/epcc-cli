#!/usr/bin/env bats
@test "pcm-products empty get collection" {
	run epcc get pcm-products
	[ $status -eq 0 ]
}

