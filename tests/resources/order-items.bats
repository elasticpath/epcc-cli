#!/usr/bin/env bats
@test "order-items empty get collection" {
	run epcc get order-items
	[ $status -eq 0 ]
}

