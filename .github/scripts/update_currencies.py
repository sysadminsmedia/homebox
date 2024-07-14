import requests
import json
import os

def fetch_currencies():
    response = requests.get('https://restcountries.com/v3.1/all')
    response.raise_for_status()
    except requests.RequestException as e:
        print(f"An error occurred: {e}")
        return []
        
    countries = response.json()

    currencies_list = []
    for country in countries:
        country_name = country.get('name', {}).get('common')
        country_currencies = country.get('currencies', {})
        for currency_code, currency_info in country_currencies.items():
            symbol = currency_info.get('symbol', '')
            # Directly use the symbol as it is
            currencies_list.append({
                'code': currency_code,
                'local': country_name,
                'symbol': symbol,
                'name': currency_info.get('name')
            })

    return currencies_list

def save_currencies(currencies, file_path):
    os.makedirs(os.path.dirname(file_path), exist_ok=True)
    with open(file_path, 'w', encoding='utf-8') as f:
        json.dump(currencies, f, ensure_ascii=False, indent=4)

def main():
    currencies = fetch_currencies()
    save_path = 'backend/internal/core/currencies/currencies.json'
    save_currencies(currencies, save_path)

if __name__ == "__main__":
    main()