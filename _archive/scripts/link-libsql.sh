#!/bin/bash
BUN_MODS=/home/enzo/rag-saldivia/node_modules/.bun
WEB_MODS=/home/enzo/rag-saldivia/apps/web/node_modules

mkdir -p $WEB_MODS/@libsql

for pkg in core hrana-client isomorphic-fetch isomorphic-ws; do
  src=$(ls $BUN_MODS | grep "@libsql+${pkg}" | head -1)
  if [ -n "$src" ]; then
    ln -sf "$BUN_MODS/$src/node_modules/@libsql/$pkg" "$WEB_MODS/@libsql/$pkg"
    echo "Linked: @libsql/$pkg"
  fi
done

# libsql native
src=$(ls $BUN_MODS | grep "^libsql@" | head -1)
if [ -n "$src" ]; then
  ln -sf "$BUN_MODS/$src/node_modules/libsql" "$WEB_MODS/libsql"
  echo "Linked: libsql"
fi

# drizzle-orm
src=$(ls $BUN_MODS | grep "^drizzle-orm@" | head -1)
if [ -n "$src" ]; then
  ln -sf "$BUN_MODS/$src/node_modules/drizzle-orm" "$WEB_MODS/drizzle-orm"
  echo "Linked: drizzle-orm"
fi

echo "All symlinks created"
ls -la $WEB_MODS/@libsql/
