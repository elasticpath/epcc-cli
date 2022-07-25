#!/usr/bin/env bats
@test "integrations empty get collection" {
	run epcc get integrations
	[ $status -eq 0 ]
}

@test "integrations delete-all support" {
	run epcc delete-all integrations
	[ $status -eq 0 ]
}


