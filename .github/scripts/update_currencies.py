import requests
import json
import os

def fetch_currencies():
    try:
        response = requests.get('https://restcountries.com/v3.1/all')
        response.raise_for_status()
    except requests.exceptions.Timeout:
        print("Request to the API timed out.")
        return []
    except requests.exceptions.RequestException as e:
        print(f"An error occurred while making the request: {e}")
        return []

    try:
        countries = response.json()
    except json.JSONDecodeError:
        print("Failed to decode JSON from the response.")
        return []

    currencies_list = []
    for country in countries:
        country_name = country.get('name', {}).get('common')
        country_currencies = country.get('currencies', {})
        for currency_code, currency_info in country_currencies.items():
            symbol = currency_info.get('symbol', '')
            currencies_list.append({
                'code': currency_code,
                'local': country_name,
                'symbol': symbol,
                'name': currency_info.get('name')
            })

    return currencies_list

def save_currencies(currencies, file_path):
    try:
        os.makedirs(os.path.dirname(file_path), exist_ok=True)
        with open(file_path, 'w', encoding='utf-8') as f:
            json.dump(currencies, f, ensure_ascii=False, indent=4)
    except IOError as e:
        print(f"An error occurred while writing to the file: {e}")

def load_existing_currencies(file_path):
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            return json.load(f)
    except (IOError, json.JSONDecodeError):
        return []  # Return an empty list if file doesn't exist or is invalid

def main():
    save_path = 'backend/internal/core/currencies/currencies.json'
    
    existing_currencies = load_existing_currencies(save_path)
    new_currencies = fetch_currencies()

    if new_currencies == existing_currencies:
        print("Currencies up-to-date with API, skipping commit.")
    else:
        save_currencies(new_currencies, save_path)
        print("Currencies updated and saved.")

if __name__ == "__main__":
    main()