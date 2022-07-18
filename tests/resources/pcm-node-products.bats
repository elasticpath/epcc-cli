#!/usr/bin/env bats
@test "pcm-node-products empty get collection" {
	run epcc get pcm-node-products
	[ $status -eq 0 ]
}

