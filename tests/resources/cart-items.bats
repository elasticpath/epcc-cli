#!/usr/bin/env bats
@test "cart-items empty get collection" {
	run epcc get cart-items
	[ $status -eq 0 ]
}

@test "cart-items delete-all support" {
	run epcc delete-all cart-items
	[ $status -eq 0 ]
}


