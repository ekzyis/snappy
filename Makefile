.PHONY: schema fixtures test

SCHEMA := client/testdata/schema.graphql
TYPEDEFS := $(wildcard stacker.news/api/typeDefs/*.js) stacker.news/lib/cursor.js
SCRIPTS := scripts/dump-schema.mjs scripts/loader.mjs scripts/register.mjs

# Regenerate the Stacker News GraphQL SDL from the git submodule.
# Only re-runs when the submodule's typeDefs (or the dump scripts) change.
schema: $(SCHEMA)

$(SCHEMA): $(TYPEDEFS) $(SCRIPTS) | scripts/node_modules
	cd scripts && npm run dump-schema

scripts/node_modules: scripts/package.json
	cd scripts && npm install --no-audit --no-fund
	touch $@

# Verify the committed fixtures still match the live API shape, and snapshot the
# raw responses under client/testdata/live/. Makes live calls to the SN API; does
# not modify the committed fixtures in client/testdata/fixtures/.
fixtures:
	go test -tags record -run TestRecordFixtures -count=1 -v ./client/

test: $(SCHEMA)
	go test ./...
