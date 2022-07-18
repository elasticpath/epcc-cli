#!/usr/bin/env bats
@test "orders empty get collection" {
	run epcc get orders
	[ $status -eq 0 ]
}

