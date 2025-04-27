#!/usr/bin/env python3
import json
import logging
import sys
from pathlib import Path

import requests
from requests.adapters import HTTPAdapter, Retry

API_URL    = 'https://restcountries.com/v3.1/all?fields=name,common,currencies'
SAVE_PATH  = Path('backend/internal/core/currencies/currencies.json')
TIMEOUT    = 10  # seconds


def setup_logging():
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s %(levelname)s: %(message)s'
    )


def fetch_currencies():
    session = requests.Session()
    retries = Retry(
        total=3,
        backoff_factor=1,
        status_forcelist=[429, 500, 502, 503, 504],
        allowed_methods=frozenset(['GET'])
    )
    session.mount('https://', HTTPAdapter(max_retries=retries))

    try:
        resp = session.get(API_URL, timeout=TIMEOUT)
        resp.raise_for_status()
    except requests.exceptions.RequestException as e:
        logging.error("API request failed: %s", e)
        return None  # signal fatal

    try:
        countries = resp.json()
    except ValueError as e:
        logging.error("Failed to parse JSON response: %s", e)
        return None  # signal fatal

    results = []
    for country in countries:
        country_name = country.get('name', {}).get('common') or "Unknown"
        for code, info in country.get('currencies', {}).items():
            results.append({
                'code':   code,
                'local':  country_name,
                'symbol': info.get('symbol', ''),
                'name':   info.get('name', '')
            })

    # sort by country name for consistency
    return sorted(results, key=lambda x: x['local'].lower())


def load_existing(path: Path):
    if not path.exists():
        return []
    try:
        with path.open('r', encoding='utf-8') as f:
            return json.load(f)
    except (IOError, ValueError) as e:
        logging.warning("Could not load existing file (%s): %s", path, e)
        return []


def save_currencies(data, path: Path):
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open('w', encoding='utf-8') as f:
        json.dump(data, f, ensure_ascii=False, indent=4)
    logging.info("Wrote %d entries to %s", len(data), path)


def main():
    setup_logging()
    logging.info("🔄 Starting currency update")

    existing = load_existing(SAVE_PATH)
    new = fetch_currencies()

    # Fatal API / parse error → exit non-zero so the workflow will fail
    if new is None:
        logging.error("Aborting: failed to fetch or parse API data.")
        sys.exit(1)

    # Empty array → log & exit zero (no file change)
    if not new:
        logging.warning("API returned empty list; skipping write.")
        sys.exit(0)

    # Identical to existing → nothing to do
    if new == existing:
        logging.info("Up-to-date; skipping write.")
        sys.exit(0)

    # Otherwise actually overwrite
    save_currencies(new, SAVE_PATH)
    logging.info("✅ Currency file updated.")


if __name__ == "__main__":
    main()
