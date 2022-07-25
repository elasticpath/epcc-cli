#!/usr/bin/env bats
@test "carts empty get collection" {
	run epcc get carts
	[ $status -eq 0 ]
}

@test "carts delete-all support" {
	run epcc delete-all carts
	[ $status -eq 0 ]
}


