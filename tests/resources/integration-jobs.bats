#!/usr/bin/env bats
@test "integration-jobs empty get collection" {
	run epcc get integration-jobs
	[ $status -eq 0 ]
}



