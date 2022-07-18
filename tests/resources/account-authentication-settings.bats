#!/usr/bin/env bats
@test "account-authentication-settings empty get collection" {
	run epcc get account-authentication-settings
	[ $status -eq 0 ]
}

