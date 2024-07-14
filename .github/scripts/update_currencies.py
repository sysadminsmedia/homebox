import requests
import json
import os

def fetch_currencies():
    try:
        response = requests.get('https://restcountries.com/v3.1/all')
        response.raise_for_status()  # Raise an error for HTTP errors
    except requests.exceptions.Timeout:
        print("Request to the API timed out.")
        return []
    except requests.exceptions.RequestException as e:
        print(f"An error occurred while making the request: {e}")
        return []

    try:
        countries = response.json()  # Attempt to parse the JSON response
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
        os.makedirs(os.path.dirname(file_path), exist_ok=True)  # Create directories if they don't exist
        with open(file_path, 'w', encoding='utf-8') as f:
            json.dump(currencies, f, ensure_ascii=False, indent=4)
    except IOError as e:
        print(f"An error occurred while writing to the file: {e}")

def main():
    currencies = fetch_currencies()
    if currencies:  # Check if currencies were successfully fetched
        save_path = 'backend/internal/core/currencies/currencies.json'
        save_currencies(currencies, save_path)

if __name__ == "__main__":
    main()
