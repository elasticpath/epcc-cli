# Design Decisions


## Loose OpenAPI Dependency

When we started integration the OpenAPI specs, there were a couple paths we could go about:

1. Leverage the OpenAPI specs heavily and require them for the tool to work.
2. Leverage the OpenAPI specs loosely and require them for tests to pass, by respecifying things in the existing yaml.

Chose 2 was chosen for a few reasons:

1. Editing the specs can be a slower process, and there are more consumers of the spec, than this tool. 
2. I suspect, that parsing and processing openapi specs every time the command starts up would be slow.

Consequently, for the most part, we don't really want to leverage the OpenAPI specs "most" of the time,
in theory someone should be able to check out the code and build the cli without the specs, with only minimal 
loss of functionality.

### Including specs in the repo.

We eventually started including specs in the repo, this was because using this tool as a library was awkward. 
