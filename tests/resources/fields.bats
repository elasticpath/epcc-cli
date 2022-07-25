#!/usr/bin/env bats
@test "fields empty get collection" {
	run epcc get fields
	[ $status -eq 0 ]
}

@test "fields delete-all support" {
	run epcc delete-all fields
	[ $status -eq 0 ]
}


