#!/usr/bin/env bats
@test "flows empty get collection" {
	run epcc get flows
	[ $status -eq 0 ]
}

@test "flows delete-all support" {
	run epcc delete-all flows
	[ $status -eq 0 ]
}


