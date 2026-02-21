#!/usr/bin/env python3
import csv
import io
import json
import logging
import os
import sys
from pathlib import Path

import requests
from requests.adapters import HTTPAdapter
from urllib3.util.retry import Retry

API_URL    = 'https://restcountries.com/v3.1/all?fields=name,common,currencies'
# Default to a pinned commit for supply-chain security
DEFAULT_ISO_4217_URL = 'https://raw.githubusercontent.com/datasets/currency-codes/052b3088938ba32028a14e75040c286c5e142145/data/codes-all.csv'
ISO_4217_URL = os.environ.get('ISO_4217_URL', DEFAULT_ISO_4217_URL)
SAVE_PATH  = Path('backend/internal/core/currencies/currencies.json')
TIMEOUT    = 10  # seconds

# Known currency decimal overrides
CURRENCY_DECIMAL_OVERRIDES = {
    "BTC": 8,  # Bitcoin uses 8 decimal places
    "JPY": 0,  # Japanese Yen has no decimal places
    "BHD": 3,  # Bahraini Dinar uses 3 decimal places
}
DEFAULT_DECIMALS = 2
MIN_DECIMALS = 0
MAX_DECIMALS = 6


def setup_logging():
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s %(levelname)s: %(message)s'
    )


def get_currency_decimals(code, iso_data):
    """
    Get the decimal places for a currency code.
    Checks overrides first, then ISO data, then uses default.
    Clamps result to safe range [MIN_DECIMALS, MAX_DECIMALS].
    """
    # Normalize the input code
    normalized_code = (code or "").strip().upper()
    
    # First check overrides
    if normalized_code in CURRENCY_DECIMAL_OVERRIDES:
        decimals = CURRENCY_DECIMAL_OVERRIDES[normalized_code]
    # Then check ISO data
    elif normalized_code in iso_data:
        decimals = iso_data[normalized_code]
    # Finally use default
    else:
        decimals = DEFAULT_DECIMALS
    
    # Ensure it's an integer and clamp to safe range
    try:
        decimals = int(decimals)
    except (ValueError, TypeError):
        decimals = DEFAULT_DECIMALS
    
    return max(MIN_DECIMALS, min(MAX_DECIMALS, decimals))


def fetch_iso_4217_data():
    """
    Fetch ISO 4217 currency data to get minor units (decimal places).
    Returns a dict mapping currency code to minor units.
    """
    # Log the resolved URL for transparency
    logging.info("Fetching ISO 4217 data from: %s", ISO_4217_URL)
    if not ISO_4217_URL.lower().startswith("https://"):
        logging.error("Refusing non-HTTPS ISO_4217_URL: %s", ISO_4217_URL)
        return {}
    
    session = requests.Session()
    retries = Retry(
        total=3,
        backoff_factor=1,
        status_forcelist=[429, 500, 502, 503, 504],
        allowed_methods=frozenset(['GET'])
    )
    session.mount('https://', HTTPAdapter(max_retries=retries))

    try:
        # Add Accept header for CSV content
        headers = {'Accept': 'text/csv'}
        resp = session.get(ISO_4217_URL, timeout=TIMEOUT, headers=headers)
        resp.raise_for_status()
    except requests.exceptions.RequestException as e:
        logging.error("Failed to fetch ISO 4217 data: %s", e)
        return {}

    # Parse CSV data
    iso_data = {}
    try:
        # Decode with utf-8-sig to strip BOM if present
        csv_content = resp.content.decode('utf-8-sig')
        csv_reader = csv.DictReader(io.StringIO(csv_content))
        
        for row in csv_reader:
            code = row.get('AlphabeticCode', '').strip()
            minor_unit = row.get('MinorUnit', '').strip()
            
            if code and minor_unit != 'N.A.':
                try:
                    # Convert minor unit to int (decimal places)
                    iso_data[code] = int(minor_unit) if minor_unit.isdigit() else 2
                except (ValueError, TypeError):
                    iso_data[code] = 2  # Default to 2 if parsing fails
                    
        logging.info("Successfully loaded decimal data for %d currencies from ISO 4217", len(iso_data))
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
            # Get decimal places using the helper function
            decimals = get_currency_decimals(code, iso_data)
            
            # Capitalize the first letter of the currency name
            currency_name = info.get('name', '')
            if currency_name:
                currency_name = currency_name[0].upper() + currency_name[1:]

            results.append({
                'code':     code,
                'local':    country_name,
                'symbol':   info.get('symbol', ''),
                'name':     currency_name,
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
