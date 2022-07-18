#!/usr/bin/env bats
@test "settings empty get collection" {
	run epcc get settings
	[ $status -eq 0 ]
}

