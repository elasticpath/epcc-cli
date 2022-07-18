#!/usr/bin/env bats
@test "files empty get collection" {
	run epcc get files
	[ $status -eq 0 ]
}

