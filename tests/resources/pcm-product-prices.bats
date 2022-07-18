#!/usr/bin/env bats
@test "pcm-product-prices empty get collection" {
	run epcc get pcm-product-prices
	[ $status -eq 0 ]
}

