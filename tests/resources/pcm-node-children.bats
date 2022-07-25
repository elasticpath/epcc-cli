#!/usr/bin/env bats
@test "pcm-node-children empty get collection" {
	run epcc get pcm-node-children
	[ $status -eq 0 ]
}



