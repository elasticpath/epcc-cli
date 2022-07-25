#!/usr/bin/env bats
@test "v2-products empty get collection" {
	run epcc get v2-products
	[ $status -eq 0 ]
}

@test "v2-products delete-all support" {
	run epcc delete-all v2-products
	[ $status -eq 0 ]
}


