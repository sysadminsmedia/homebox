import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { fmtDate } from './use-formatters';

// Mock required functions globally
(global as any).validDate = vi.fn(() => true);
(global as any).useNuxtApp = vi.fn(() => ({
  $i18nGlobal: {
    locale: {
      value: 'en-US'
    }
  }
}));
(global as any).useViewPreferences = vi.fn(() => ({
  value: {
    overrideFormatLocale: null
  }
}));
(global as any).useRuntimeConfig = vi.fn(() => mockRuntimeConfig);

const mockRuntimeConfig = {
  public: {
    hboxDateFormatHuman: '',
    hboxDateFormatLong: '',
    hboxDateFormatShort: '',
  }
};

vi.mock('#app', () => ({
  useRuntimeConfig: () => mockRuntimeConfig,
  useNuxtApp: () => ({
    $i18nGlobal: {
      locale: {
        value: 'en-US'
      }
    }
  })
}));

vi.mock('~/composables/utils', () => ({
  validDate: vi.fn(() => true)
}));

vi.mock('~/composables/use-preferences', () => ({
  useViewPreferences: () => ({
    value: {
      overrideFormatLocale: null
    }
  })
}));

vi.mock('date-fns', () => ({
  format: vi.fn((date, formatStr) => {
    if (formatStr === 'PPP') return 'January 2nd, 2006';
    if (formatStr === 'PP') return 'Jan 2, 2006';
    if (formatStr === 'P') return '01/02/2006';
    if (formatStr === 'dd/MM/yyyy') return '02/01/2006';
    if (formatStr === 'yyyy-MM-dd') return '2006-01-02';
    if (formatStr === 'MMM d, yyyy') return 'Jan 2, 2006';
    return formatStr; // fallback
  })
}));

describe('fmtDate with environment variable overrides', () => {
  const testDate = new Date('2006-01-02T15:04:05-07:00');

  beforeEach(() => {
    mockRuntimeConfig.public.hboxDateFormatHuman = '';
    mockRuntimeConfig.public.hboxDateFormatLong = '';
    mockRuntimeConfig.public.hboxDateFormatShort = '';
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('default behavior (no environment variables)', () => {
    it('should use default PPP format for human', () => {
      const result = fmtDate(testDate, 'human');
      expect(result).toBe('January 2nd, 2006');
    });

    it('should use default PP format for long', () => {
      const result = fmtDate(testDate, 'long');
      expect(result).toBe('Jan 2, 2006');
    });

    it('should use default P format for short', () => {
      const result = fmtDate(testDate, 'short');
      expect(result).toBe('01/02/2006');
    });
  });

  describe('with environment variable overrides', () => {
    it('should use custom format for human when HBOX_DATE_FORMAT_HUMAN is set', () => {
      mockRuntimeConfig.public.hboxDateFormatHuman = 'MMM d, yyyy';
      const result = fmtDate(testDate, 'human');
      expect(result).toBe('Jan 2, 2006');
    });

    it('should use custom format for long when HBOX_DATE_FORMAT_LONG is set', () => {
      mockRuntimeConfig.public.hboxDateFormatLong = 'yyyy-MM-dd';
      const result = fmtDate(testDate, 'long');
      expect(result).toBe('2006-01-02');
    });

    it('should use custom format for short when HBOX_DATE_FORMAT_SHORT is set', () => {
      mockRuntimeConfig.public.hboxDateFormatShort = 'dd/MM/yyyy';
      const result = fmtDate(testDate, 'short');
      expect(result).toBe('02/01/2006');
    });

    it('should fallback to default when custom format is empty string', () => {
      mockRuntimeConfig.public.hboxDateFormatShort = '';
      const result = fmtDate(testDate, 'short');
      expect(result).toBe('01/02/2006');
    });
  });

  describe('edge cases', () => {
    it('should return empty string for invalid format type', () => {
      const result = fmtDate(testDate, 'invalid' as any);
      expect(result).toBe('');
    });

    it('should handle string date input', () => {
      const result = fmtDate('2006-01-02', 'short');
      expect(result).toBe('01/02/2006');
    });

    it('should handle number date input', () => {
      const result = fmtDate(1136239445000, 'short'); 
      expect(result).toBe('01/02/2006');
    });
  });
});