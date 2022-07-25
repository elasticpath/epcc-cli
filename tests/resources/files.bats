#!/usr/bin/env bats
@test "files empty get collection" {
	run epcc get files
	[ $status -eq 0 ]
}

@test "files delete-all support" {
	run epcc delete-all files
	[ $status -eq 0 ]
}


