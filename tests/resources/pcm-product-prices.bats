#!/usr/bin/env bats
@test "pcm-product-prices empty get collection" {
	run epcc get pcm-product-prices
	[ $status -eq 0 ]
}

@test "pcm-product-prices delete-all support" {
	run epcc delete-all pcm-product-prices
	[ $status -eq 0 ]
}


