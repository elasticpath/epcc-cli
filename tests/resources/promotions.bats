#!/usr/bin/env bats
@test "promotions empty get collection" {
	run epcc get promotions
	[ $status -eq 0 ]
}

@test "promotions delete-all support" {
	run epcc delete-all promotions
	[ $status -eq 0 ]
}


