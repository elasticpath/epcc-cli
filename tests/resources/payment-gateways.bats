#!/usr/bin/env bats
@test "payment-gateways empty get collection" {
	run epcc get payment-gateways
	[ $status -eq 0 ]
}



