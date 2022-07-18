#!/usr/bin/env bats
@test "cart-items empty get collection" {
	run epcc get cart-items
	[ $status -eq 0 ]
}

