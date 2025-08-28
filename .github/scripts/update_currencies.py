#!/usr/bin/env python3
import csv
import io
import json
import logging
import sys
from pathlib import Path

import requests
from requests.adapters import HTTPAdapter, Retry

API_URL    = 'https://restcountries.com/v3.1/all?fields=name,common,currencies'
ISO_4217_URL = 'https://raw.githubusercontent.com/datasets/currency-codes/master/data/codes-all.csv'
SAVE_PATH  = Path('backend/internal/core/currencies/currencies.json')
TIMEOUT    = 10  # seconds


def setup_logging():
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s %(levelname)s: %(message)s'
    )


def fetch_iso_4217_data():
    """
    Fetch ISO 4217 currency data to get minor units (decimal places).
    Returns a dict mapping currency code to minor units.
    """
    session = requests.Session()
    retries = Retry(
        total=3,
        backoff_factor=1,
        status_forcelist=[429, 500, 502, 503, 504],
        allowed_methods=frozenset(['GET'])
    )
    session.mount('https://', HTTPAdapter(max_retries=retries))

    try:
        resp = session.get(ISO_4217_URL, timeout=TIMEOUT)
        resp.raise_for_status()
    except requests.exceptions.RequestException as e:
        logging.error("Failed to fetch ISO 4217 data: %s", e)
        return {}

    # Parse CSV data
    iso_data = {}
    try:
        csv_reader = csv.DictReader(io.StringIO(resp.text))
        for row in csv_reader:
            code = row.get('AlphabeticCode', '').strip()
            minor_unit = row.get('MinorUnit', '').strip()
            
            if code and minor_unit != 'N.A.':
                try:
                    # Convert minor unit to int (decimal places)
                    iso_data[code] = int(minor_unit) if minor_unit.isdigit() else 2
                except (ValueError, TypeError):
                    iso_data[code] = 2  # Default to 2 if parsing fails
                    
        logging.info("Loaded decimal data for %d currencies from ISO 4217", len(iso_data))
        return iso_data
        
    except Exception as e:
        logging.error("Failed to parse ISO 4217 CSV data: %s", e)
        return {}


def fetch_currencies():
    # First, fetch ISO 4217 data for decimal places
    iso_data = fetch_iso_4217_data()
    
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
            # Get decimal places from ISO 4217 data, default to 2 if not found
            decimals = iso_data.get(code, 2)
            
            results.append({
                'code':     code,
                'local':    country_name,
                'symbol':   info.get('symbol', ''),
                'name':     info.get('name', ''),
                'decimals': decimals
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
