#!/usr/bin/env python3
"""
Script to automatically update language names in the English translation file.
Scans for locale files and ensures all languages are present in en.json.
"""
import json
import logging
import sys
from pathlib import Path
from typing import Dict, List

from babel import Locale, UnknownLocaleError

LOCALES_DIR = Path('frontend/locales')
EN_JSON_PATH = LOCALES_DIR / 'en.json'

# Mapping for special/custom locale codes that don't follow standard BCP 47
CUSTOM_LOCALE_MAPPINGS = {
    'ar-AA': ('ar', 'Arabic'),  # Generic Arabic
    'en': ('en', 'English'),
}


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
        locale_codes.append(locale_code)
    
    logging.info("Found %d locale files", len(locale_codes))
    return sorted(locale_codes)


def get_language_name(locale_code: str) -> str:
    """
    Get the English display name for a locale code.
    
    Args:
        locale_code: Language/locale code (e.g., 'en', 'pt-BR', 'zh-CN')
    
    Returns:
        English display name for the language
    """
    # Check custom mappings first
    if locale_code in CUSTOM_LOCALE_MAPPINGS:
        _, name = CUSTOM_LOCALE_MAPPINGS[locale_code]
        return name
    
    try:
        # Parse locale code using Babel
        locale = Locale.parse(locale_code.replace('-', '_'))
        
        # Get English display name
        display_name = locale.get_display_name('en')
        
        if not display_name:
            # Fallback to language name if full display name not available
            display_name = locale.english_name
        
        return display_name
    
    except (UnknownLocaleError, ValueError) as e:
        logging.warning("Could not parse locale '%s': %s", locale_code, e)
        # Fallback: capitalize the locale code
        return locale_code.replace('-', ' ').replace('_', ' ').title()


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


def update_language_names(en_data: dict, locale_codes: List[str]) -> bool:
    """
    Update the languages section in en.json with all locale codes.
    
    Args:
        en_data: The parsed en.json data
        locale_codes: List of all locale codes from files
    
    Returns:
        True if changes were made, False otherwise
    """
    # Ensure languages section exists
    if 'languages' not in en_data:
        en_data['languages'] = {}
        logging.info("Created 'languages' section in en.json")
    
    languages = en_data['languages']
    original_languages = languages.copy()
    
    # Add any missing languages
    added_count = 0
    for locale_code in locale_codes:
        if locale_code not in languages:
            language_name = get_language_name(locale_code)
            languages[locale_code] = language_name
            logging.info("Added language: %s = %s", locale_code, language_name)
            added_count += 1
    
    # Sort languages alphabetically by key
    en_data['languages'] = dict(sorted(languages.items()))
    
    # Check if anything changed
    changed = (original_languages != en_data['languages'])
    
    if changed:
        logging.info("Updated %d language names", added_count)
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
    
    # Update language names
    changed = update_language_names(en_data, locale_codes)
    
    if changed:
        save_en_json(en_data)
        logging.info("✅ Language names updated successfully")
    else:
        logging.info("✅ No updates needed, en.json is already up-to-date")
    
    sys.exit(0)


if __name__ == "__main__":
    main()
