#!/usr/bin/env bats
@test "promotion-codes empty get collection" {
	run epcc get promotion-codes
	[ $status -eq 0 ]
}

@test "promotion-codes delete-all support" {
	run epcc delete-all promotion-codes
	[ $status -eq 0 ]
}


