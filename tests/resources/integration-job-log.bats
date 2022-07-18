#!/usr/bin/env bats
@test "integration-job-log empty get collection" {
	run epcc get integration-job-log
	[ $status -eq 0 ]
}

