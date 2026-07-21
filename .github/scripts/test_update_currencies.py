#!/usr/bin/env python3
import unittest
from unittest import mock

import update_currencies


class ParseCountryPageTests(unittest.TestCase):
    def test_parses_v5_response(self):
        countries, meta = update_currencies.parse_country_page({
            'data': {
                'objects': [{'names': {'common': 'Canada'}, 'currencies': []}],
                'meta': {'more': False, 'total': 1},
            }
        })

        self.assertEqual('Canada', countries[0]['names']['common'])
        self.assertFalse(meta['more'])

    def test_reports_api_error_message(self):
        with self.assertRaisesRegex(ValueError, 'API key is invalid'):
            update_currencies.parse_country_page({
                'errors': [{'message': 'API key is invalid'}]
            })

    def test_rejects_legacy_response(self):
        with self.assertRaisesRegex(ValueError, 'missing the data object'):
            update_currencies.parse_country_page({
                'success': False,
                'errors': [],
            })

    def test_requires_complete_pagination_metadata(self):
        with self.assertRaisesRegex(ValueError, 'total count'):
            update_currencies.parse_country_page({
                'data': {
                    'objects': [],
                    'meta': {'more': False},
                }
            })


class CountriesToCurrenciesTests(unittest.TestCase):
    def test_converts_v5_currency_list(self):
        result = update_currencies.countries_to_currencies([
            {
                'names': {'common': 'Canada'},
                'currencies': [{
                    'code': 'CAD',
                    'name': 'Canadian dollar',
                    'symbol': '$',
                }],
            }
        ], {'CAD': 2})

        self.assertEqual([{
            'code': 'CAD',
            'local': 'Canada',
            'symbol': '$',
            'name': 'Canadian dollar',
            'decimals': 2,
        }], result)

    def test_rejects_old_currency_object_shape(self):
        with self.assertRaisesRegex(ValueError, 'currencies must be a JSON list'):
            update_currencies.countries_to_currencies([
                {
                    'names': {'common': 'Canada'},
                    'currencies': {'CAD': {'name': 'Canadian dollar'}},
                }
            ], {})


class FetchCurrenciesTests(unittest.TestCase):
    class FakeResponse:
        def __init__(self, payload):
            self.payload = payload

        def raise_for_status(self):
            return None

        def json(self):
            return self.payload

    class FakeSession:
        def __init__(self, payloads):
            self.payloads = iter(payloads)
            self.calls = []

        def get(self, url, **kwargs):
            self.calls.append((url, kwargs))
            return FetchCurrenciesTests.FakeResponse(next(self.payloads))

    def test_fetches_every_page_with_authentication(self):
        pages = [
            {
                'data': {
                    'objects': [{
                        'names': {'common': 'Canada'},
                        'currencies': [{
                            'code': 'CAD',
                            'name': 'Canadian dollar',
                            'symbol': '$',
                        }],
                    }],
                    'meta': {'more': True, 'total': 2},
                }
            },
            {
                'data': {
                    'objects': [{
                        'names': {'common': 'Japan'},
                        'currencies': [{
                            'code': 'JPY',
                            'name': 'Japanese yen',
                            'symbol': '¥',
                        }],
                    }],
                    'meta': {'more': False, 'total': 2},
                }
            },
        ]
        session = self.FakeSession(pages)

        with mock.patch.dict(update_currencies.os.environ, {
            update_currencies.API_KEY_ENV: 'test-key',
        }), mock.patch.object(update_currencies, 'create_session', return_value=session), \
                mock.patch.object(update_currencies, 'fetch_iso_4217_data', return_value={'CAD': 2}):
            result = update_currencies.fetch_currencies()

        self.assertEqual(['CAD', 'JPY'], [currency['code'] for currency in result])
        self.assertEqual(2, len(session.calls))
        self.assertEqual('Bearer test-key', session.calls[0][1]['headers']['Authorization'])
        self.assertEqual(0, session.calls[0][1]['params']['offset'])
        self.assertEqual(1, session.calls[1][1]['params']['offset'])

    def test_missing_api_key_is_fatal_before_network_access(self):
        with mock.patch.dict(update_currencies.os.environ, {}, clear=True), \
                mock.patch.object(update_currencies, 'create_session') as create_session:
            result = update_currencies.fetch_currencies()

        self.assertIsNone(result)
        create_session.assert_not_called()


if __name__ == '__main__':
    unittest.main()
