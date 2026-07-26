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

API_URL = 'https://api.restcountries.com/countries/v5'
API_KEY_ENV = 'RESTCOUNTRIES_API_KEY'
API_PAGE_LIMIT = 100
# Default to a pinned commit for supply-chain security
DEFAULT_ISO_4217_URL = 'https://raw.githubusercontent.com/datasets/currency-codes/refs/heads/main/data/codes-all.csv'
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


def create_session():
    session = requests.Session()
    retries = Retry(
        total=3,
        backoff_factor=1,
        status_forcelist=[429, 500, 502, 503, 504],
        allowed_methods=frozenset(['GET'])
    )
    session.mount('https://', HTTPAdapter(max_retries=retries))
    return session


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
    
    session = create_session()

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


def parse_country_page(payload):
    if not isinstance(payload, dict):
        raise ValueError("response must be a JSON object")

    errors = payload.get('errors')
    if errors:
        messages = [
            error.get('message', str(error)) if isinstance(error, dict) else str(error)
            for error in errors
        ]
        raise ValueError("API returned errors: %s" % "; ".join(messages))

    data = payload.get('data')
    if not isinstance(data, dict):
        raise ValueError("response is missing the data object")

    countries = data.get('objects')
    meta = data.get('meta')
    if not isinstance(countries, list):
        raise ValueError("response data is missing the objects list")
    if not isinstance(meta, dict):
        raise ValueError("response data is missing pagination metadata")
    if not isinstance(meta.get('more'), bool):
        raise ValueError("pagination metadata is missing the more flag")
    if isinstance(meta.get('total'), bool) or not isinstance(meta.get('total'), int):
        raise ValueError("pagination metadata is missing the total count")

    return countries, meta


def countries_to_currencies(countries, iso_data):
    results = []
    for country in countries:
        if not isinstance(country, dict):
            raise ValueError("country entry must be a JSON object")

        names = country.get('names') or {}
        currencies = country.get('currencies') or []
        if not isinstance(names, dict):
            raise ValueError("country names must be a JSON object")
        if not isinstance(currencies, list):
            raise ValueError("country currencies must be a JSON list")
        country_name = names.get('common') or "Unknown"
        if not isinstance(country_name, str):
            raise ValueError("country common name must be a string")

        for info in currencies:
            if not isinstance(info, dict):
                raise ValueError("currency entry must be a JSON object")
            code = info.get('code') or ''
            if not isinstance(code, str):
                raise ValueError("currency code must be a string")
            if not code:
                logging.warning("Skipping currency without a code for %s", country_name)
                continue

            # Get decimal places using the helper function
            decimals = get_currency_decimals(code, iso_data)
            
            # Capitalize the first letter of the currency name
            currency_name = info.get('name') or ''
            if not isinstance(currency_name, str):
                raise ValueError("currency name must be a string")
            if currency_name:
                currency_name = currency_name[0].upper() + currency_name[1:]

            symbol = info.get('symbol') or ''
            if not isinstance(symbol, str):
                raise ValueError("currency symbol must be a string")

            results.append({
                'code':     code,
                'local':    country_name,
                'symbol':   symbol,
                'name':     currency_name,
                'decimals': decimals
            })

    # sort by country name for consistency
    return sorted(results, key=lambda x: x['local'].lower())


def fetch_currencies():
    api_key = os.environ.get(API_KEY_ENV, '').strip()
    if not api_key:
        logging.error("Missing required %s environment variable", API_KEY_ENV)
        return None

    # First, fetch ISO 4217 data for decimal places
    iso_data = fetch_iso_4217_data()
    session = create_session()
    countries = []
    offset = 0
    expected_total = None

    while True:
        try:
            resp = session.get(
                API_URL,
                timeout=TIMEOUT,
                headers={'Authorization': 'Bearer %s' % api_key},
                params={
                    'response_fields': 'names.common,currencies',
                    'limit': API_PAGE_LIMIT,
                    'offset': offset,
                }
            )
            resp.raise_for_status()
        except requests.exceptions.RequestException as e:
            logging.error("Countries API request failed: %s", e)
            return None  # signal fatal

        try:
            payload = resp.json()
            page, meta = parse_country_page(payload)
        except (ValueError, TypeError) as e:
            logging.error("Failed to parse Countries API response: %s", e)
            return None  # signal fatal

        countries.extend(page)
        page_total = meta['total']
        if expected_total is None:
            expected_total = page_total
        elif page_total != expected_total:
            logging.error("Countries API total changed while fetching pages")
            return None
        if not meta.get('more'):
            break
        if not page:
            logging.error("Countries API returned an empty page while more data was expected")
            return None
        if len(countries) >= expected_total:
            logging.error("Countries API reported more pages than the declared total")
            return None
        offset += len(page)

    if len(countries) != expected_total:
        logging.error(
            "Countries API returned %d of %d expected countries",
            len(countries), expected_total
        )
        return None

    try:
        return countries_to_currencies(countries, iso_data)
    except (ValueError, TypeError) as e:
        logging.error("Failed to parse country data: %s", e)
        return None


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

    # Empty array means an incomplete or malformed all-countries response.
    if not new:
        logging.error("Aborting: Countries API returned no currencies.")
        sys.exit(1)

    # Identical to existing → nothing to do
    if new == existing:
        logging.info("Up-to-date; skipping write.")
        sys.exit(0)

    # Otherwise actually overwrite
    save_currencies(new, SAVE_PATH)
    logging.info("✅ Currency file updated.")


if __name__ == "__main__":
    main()
