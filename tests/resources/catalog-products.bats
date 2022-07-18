#!/usr/bin/env bats
@test "catalog-products empty get collection" {
	run epcc get catalog-products
	[ $status -eq 0 ]
}

