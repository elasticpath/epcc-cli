#!/usr/bin/env bats
@test "currencies empty get collection" {
	run epcc get currencies
	[ $status -eq 0 ]
}

@test "currencies delete-all support" {
	run epcc delete-all currencies
	[ $status -eq 0 ]
}


