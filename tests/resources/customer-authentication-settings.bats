#!/usr/bin/env bats
@test "customer-authentication-settings empty get collection" {
	run epcc get customer-authentication-settings
	[ $status -eq 0 ]
}



