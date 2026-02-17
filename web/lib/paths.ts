import { resolve } from "path"

const PROJECT_ROOT = resolve(process.cwd(), "..")

export const CONFIG_DIR = resolve(PROJECT_ROOT, "config")
export const DATA_DIR = resolve(PROJECT_ROOT, "data")
export const OUTPUT_DIR = resolve(PROJECT_ROOT, "output")

export const configPaths = {
  settings: resolve(CONFIG_DIR, "settings.json"),
  settingsSchema: resolve(CONFIG_DIR, "settings.schema.json"),
  watchlist: resolve(CONFIG_DIR, "watchlist.json"),
  watchlistSchema: resolve(CONFIG_DIR, "watchlist.schema.json"),
  holdings: resolve(CONFIG_DIR, "holdings.json"),
  holdingsSchema: resolve(CONFIG_DIR, "holdings.schema.json"),
  theses: resolve(CONFIG_DIR, "theses.md"),
}
