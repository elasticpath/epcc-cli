#!/usr/bin/env bats
@test "v2-products empty get collection" {
	run epcc get v2-products
	[ $status -eq 0 ]
}

