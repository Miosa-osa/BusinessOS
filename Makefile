# Optimal Engine — Make targets
#
# make install        install deps + compile
# make bootstrap      migrate + ingest sample-workspace/
# make dev            boot the engine with API enabled
# make ui             boot the desktop dev server
# make test           full test suite
# make reality        run the reality-check probes
# make clean          wipe _build/ and the dev SQLite

.PHONY: install bootstrap dev ui test reality clean seed help

help:
	@echo "Optimal Engine — make targets"
	@echo ""
	@echo "  make install      — mix deps.get + mix compile"
	@echo "  make bootstrap    — compile, migrate, ingest sample-workspace/"
	@echo "  make dev          — iex -S mix (engine + HTTP API)"
	@echo "  make ui           — desktop: npm install + vite dev"
	@echo "  make test         — full test suite"
	@echo "  make reality      — mix optimal.reality_check --hard"
	@echo "  make clean        — wipe _build/ and the dev SQLite"
	@echo ""
	@echo "  Quick start:  make install && make bootstrap && make dev"

install:
	mix deps.get
	mix compile

bootstrap: install
	mix optimal.bootstrap

seed:
	mix optimal.bootstrap

dev:
	iex -S mix

ui:
	cd desktop && npm install && npm run dev

test:
	mix test

reality:
	mix optimal.reality_check --hard

clean:
	rm -rf _build/ .optimal/index.db* deps/
