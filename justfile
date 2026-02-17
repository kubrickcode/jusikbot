set dotenv-load := true

root_dir := justfile_directory()

deps: deps-root deps-web

deps-root:
    pnpm install

deps-web:
    cd {{ root_dir }}/web && pnpm install

install-psql:
    #!/usr/bin/env bash
    set -euo pipefail
    if command -v psql &> /dev/null; then
      echo "psql already installed: $(psql --version)"
    else
      DEBIAN_FRONTEND=noninteractive apt-get update && \
        apt-get -y install lsb-release wget && \
        wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add - && \
        echo "deb http://apt.postgresql.org/pub/repos/apt/ $(lsb_release -cs)-pgdg main" | tee /etc/apt/sources.list.d/pgdg.list && \
        apt-get update && \
        apt-get -y install postgresql-client-17
    fi

kill-port port:
    #!/usr/bin/env bash
    set -euo pipefail
    pid=$(ss -tlnp | grep ":{{ port }} " | sed -n 's/.*pid=\([0-9]*\).*/\1/p' | head -1)
    if [ -n "$pid" ]; then
        echo "Killing process $pid on port {{ port }}"
        kill -9 $pid
    else
        echo "No process found on port {{ port }}"
    fi

migrate:
    cd {{ root_dir }}/collector && go run ./cmd/migrate

reset:
    #!/usr/bin/env bash
    set -euo pipefail
    psql "$DATABASE_URL" -c "DROP TABLE IF EXISTS price_history, fx_rate, schema_version CASCADE;"
    echo "tables dropped"
    just migrate

collect target='all':
    cd {{ root_dir }}/collector && go run ./cmd/collect --target {{ target }}

dev-web:
    cd {{ root_dir }}/web && pnpm dev

test-collector:
    cd {{ root_dir }}/collector && go test ./...

lint target="all":
    #!/usr/bin/env bash
    set -euox pipefail
    case "{{ target }}" in
      all)
        just lint justfile
        just lint config
        ;;
      justfile)
        just --fmt --unstable
        ;;
      config)
        npx prettier --write "**/*.{json,yml,yaml,md}"
        ;;
      *)
        echo "Unknown target: {{ target }}"
        exit 1
        ;;
    esac
