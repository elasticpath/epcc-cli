#!/usr/bin/env bats
@test "promotion-codes empty get collection" {
	run epcc get promotion-codes
	[ $status -eq 0 ]
}

