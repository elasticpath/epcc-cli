#!/usr/bin/env bats
@test "carts empty get collection" {
	run epcc get carts
	[ $status -eq 0 ]
}

