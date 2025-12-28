#!/usr/bin/env python3
"""
Script to automatically update language names in the English translation file.
Queries Weblate for translation completion and language names.
Only adds languages with >=80% completion to en.json.
"""
import json
import logging
import sys
from pathlib import Path
from typing import Dict, List, Optional

import requests
from babel import Locale, UnknownLocaleError

LOCALES_DIR = Path('frontend/locales')
EN_JSON_PATH = LOCALES_DIR / 'en.json'
WEBLATE_API_URL = 'https://translate.sysadminsmedia.com/api'
WEBLATE_PROJECT = 'homebox'
WEBLATE_COMPONENT = 'frontend'
COMPLETION_THRESHOLD = 80.0  # Minimum completion percentage to include language
TIMEOUT = 10  # seconds


def setup_logging():
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s %(levelname)s: %(message)s'
    )


def get_locale_files() -> List[str]:
    """Get all locale codes from JSON files in the locales directory."""
    if not LOCALES_DIR.exists():
        logging.error("Locales directory not found: %s", LOCALES_DIR)
        return []
    
    locale_codes = []
    for file in sorted(LOCALES_DIR.glob('*.json')):
        # Extract locale code from filename (e.g., "en.json" -> "en")
        locale_code = file.stem
        # Validate locale code format - should not contain dots
        if '.' not in locale_code:
            locale_codes.append(locale_code)
        else:
            logging.warning("Skipping invalid locale code: %s", locale_code)
    
    logging.info("Found %d locale files", len(locale_codes))
    return sorted(locale_codes)


def fetch_weblate_translations() -> Optional[Dict[str, Dict]]:
    """
    Fetch translation statistics from Weblate API.
    
    Returns:
        Dict mapping locale code to translation data (percent, name, native_name)
        or None if API is unavailable
    """
    url = f"{WEBLATE_API_URL}/components/{WEBLATE_PROJECT}/{WEBLATE_COMPONENT}/translations/"
    
    try:
        # Weblate API may require pagination
        translations = {}
        page_url = url
        
        while page_url:
            logging.info("Fetching translations from Weblate: %s", page_url)
            resp = requests.get(page_url, timeout=TIMEOUT)
            
            if resp.status_code != 200:
                logging.warning("Weblate API returned status %d", resp.status_code)
                return None
            
            data = resp.json()
            
            for trans in data.get('results', []):
                # Weblate uses underscores, we use hyphens
                locale_code = trans.get('language_code', '').replace('_', '-')
                percent = trans.get('translated_percent', 0.0)
                
                lang_info = trans.get('language', {})
                english_name = lang_info.get('name', '')
                native_name = lang_info.get('native', '')
                
                translations[locale_code] = {
                    'percent': percent,
                    'english_name': english_name,
                    'native_name': native_name
                }
            
            # Check for next page
            page_url = data.get('next')
        
        logging.info("Fetched %d translations from Weblate", len(translations))
        return translations
    
    except requests.exceptions.RequestException as e:
        logging.warning("Failed to fetch from Weblate API: %s", e)
        return None
    except Exception as e:
        logging.error("Unexpected error fetching Weblate data: %s", e)
        return None


def get_language_name_from_babel(locale_code: str) -> Optional[str]:
    """
    Get the language name using Babel in format "English (Native)".
    
    Args:
        locale_code: Language/locale code (e.g., 'en', 'pt-BR', 'zh-CN')
    
    Returns:
        Language name in format "English (Native)" or None if cannot parse
    """
    try:
        # Special handling for ar-AA (non-standard code, use standard 'ar')
        if locale_code == 'ar-AA':
            locale = Locale.parse('ar')
        else:
            # Parse locale code using Babel
            locale = Locale.parse(locale_code.replace('-', '_'))
        
        # Get English display name
        english_name = locale.get_display_name('en')
        
        # Get native display name
        native_name = locale.get_display_name(locale)
        
        if not english_name:
            return None
        
        # Format: "English (Native)" if native name differs and is available
        if native_name and native_name != english_name:
            # Clean up nested parentheses for complex locales
            if '(' in english_name and '(' in native_name:
                # For cases like "Japanese (Japan) (日本語 (日本))"
                # Simplify to "Japanese (日本語)"
                english_base = english_name.split('(')[0].strip()
                native_base = native_name.split('(')[0].strip()
                return f"{english_base} ({native_base})"
            else:
                return f"{english_name} ({native_name})"
        else:
            return english_name
    
    except (UnknownLocaleError, ValueError) as e:
        logging.debug("Could not parse locale '%s' with Babel: %s", locale_code, e)
        return None


def get_language_name(locale_code: str, weblate_data: Optional[Dict] = None) -> Optional[str]:
    """
    Get the display name for a locale code.
    Priority: Weblate API > Babel > None
    
    Args:
        locale_code: Language/locale code (e.g., 'en', 'pt-BR', 'zh-CN')
        weblate_data: Translation data from Weblate (if available)
    
    Returns:
        Language name in format "English (Native)" or None if invalid
    """
    # Validate locale code format
    if '.' in locale_code or locale_code.startswith('languages.'):
        logging.error("Invalid locale code format: %s", locale_code)
        return None
    
    # Try Weblate first
    if weblate_data and locale_code in weblate_data:
        english_name = weblate_data[locale_code].get('english_name', '')
        native_name = weblate_data[locale_code].get('native_name', '')
        
        if english_name:
            # Format: "English (Native)" if both names available and different
            if native_name and native_name != english_name:
                return f"{english_name} ({native_name})"
            else:
                return english_name
    
    # Fallback to Babel
    babel_name = get_language_name_from_babel(locale_code)
    if babel_name:
        return babel_name
    
    # If all else fails, return None (don't guess)
    logging.warning("Could not determine language name for: %s", locale_code)
    return None


def load_en_json() -> dict:
    """Load the English translation JSON file."""
    if not EN_JSON_PATH.exists():
        logging.error("English translation file not found: %s", EN_JSON_PATH)
        return {}
    
    try:
        with EN_JSON_PATH.open('r', encoding='utf-8') as f:
            return json.load(f)
    except (IOError, json.JSONDecodeError) as e:
        logging.error("Failed to load %s: %s", EN_JSON_PATH, e)
        return {}


def save_en_json(data: dict):
    """Save the English translation JSON file."""
    try:
        with EN_JSON_PATH.open('w', encoding='utf-8') as f:
            # Use 4-space indentation to match existing file format
            json.dump(data, f, ensure_ascii=False, indent=4)
            # Add newline at end of file
            f.write('\n')
        logging.info("Saved updated en.json")
    except IOError as e:
        logging.error("Failed to save %s: %s", EN_JSON_PATH, e)
        sys.exit(1)


def update_language_names(en_data: dict, locale_codes: List[str], weblate_data: Optional[Dict] = None) -> bool:
    """
    Update the languages section in en.json.
    - Add new languages with >=80% completion (from Weblate) or that exist as locale files
    - Never remove existing entries (even if completion drops below 80%)
    
    Args:
        en_data: The parsed en.json data
        locale_codes: List of all locale codes from files
        weblate_data: Translation data from Weblate (if available)
    
    Returns:
        True if changes were made, False otherwise
    """
    # Ensure languages section exists
    if 'languages' not in en_data:
        en_data['languages'] = {}
        logging.info("Created 'languages' section in en.json")
    
    languages = en_data['languages']
    original_languages = languages.copy()
    
    # Process each locale file
    added_count = 0
    skipped_count = 0
    
    for locale_code in locale_codes:
        # Skip if already in languages (never remove existing entries)
        if locale_code in languages:
            continue
        
        # Check Weblate completion threshold if data available
        if weblate_data and locale_code in weblate_data:
            percent = weblate_data[locale_code].get('percent', 0.0)
            
            if percent < COMPLETION_THRESHOLD:
                logging.info("Skipping %s: %.1f%% completion (threshold: %.1f%%)", 
                           locale_code, percent, COMPLETION_THRESHOLD)
                skipped_count += 1
                continue
            else:
                logging.info("Including %s: %.1f%% completion", locale_code, percent)
        else:
            # If Weblate data not available, include locale file but log warning
            logging.info("Including %s: Weblate data not available, locale file exists", locale_code)
        
        # Get language name
        language_name = get_language_name(locale_code, weblate_data)
        
        if language_name:
            languages[locale_code] = language_name
            logging.info("Added language: %s = %s", locale_code, language_name)
            added_count += 1
        else:
            logging.warning("Skipping %s: could not determine language name", locale_code)
            skipped_count += 1
    
    # Sort languages alphabetically by key
    en_data['languages'] = dict(sorted(languages.items()))
    
    # Check if anything changed
    changed = (original_languages != en_data['languages'])
    
    if changed:
        logging.info("Updated %d language names, skipped %d", added_count, skipped_count)
    else:
        logging.info("All languages already present, no changes needed")
    
    return changed


def main():
    setup_logging()
    logging.info("🔄 Starting language names update")
    
    # Get all locale files
    locale_codes = get_locale_files()
    if not locale_codes:
        logging.error("No locale files found")
        sys.exit(1)
    
    # Load English translation file
    en_data = load_en_json()
    if not en_data:
        logging.error("Failed to load English translation file")
        sys.exit(1)
    
    # Fetch Weblate translation statistics
    weblate_data = fetch_weblate_translations()
    if weblate_data:
        logging.info("Successfully fetched Weblate data for %d languages", len(weblate_data))
    else:
        logging.warning("Weblate data not available, proceeding with locale files only")
    
    # Update language names
    changed = update_language_names(en_data, locale_codes, weblate_data)
    
    if changed:
        save_en_json(en_data)
        logging.info("✅ Language names updated successfully")
    else:
        logging.info("✅ No updates needed, en.json is already up-to-date")
    
    sys.exit(0)


if __name__ == "__main__":
    main()
